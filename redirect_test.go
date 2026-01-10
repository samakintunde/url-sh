package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"
	"url-shortener/tests"
)

func TestRedirectAndAnalytics(t *testing.T) {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	t.Cleanup(cancel)

	cfg := tests.BuildTestConfig(8102)
	go run(ctx, cfg)

	if err := tests.WaitForReady(ctx, 10*time.Second, fmt.Sprintf("http://localhost:%d/health", cfg.Server.Port)); err != nil {
		t.Fatalf("server not ready: %v", err)
	}

	email := "redirecter@example.com"
	password := "SecurePass123."
	cookie := createUserAndLogin(t, ctx, cfg, email, password)

	// Create Link
	originalURL := "https://example.org"
	prettyID := "my-pretty-link"
	createBody := fmt.Sprintf(`{"original_url": "%s", "pretty_id": "%s"}`, originalURL, prettyID)
	resp, err := tests.DoRequestWithCookie(ctx, http.MethodPost, tests.BuildRequestUrl(cfg.Server, "/api/links"), bytes.NewReader([]byte(createBody)), cookie)
	if err != nil || resp.StatusCode != http.StatusCreated {
		t.Fatalf("create pretty link failed")
	}

	var createResp struct {
		Data struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	json.NewDecoder(resp.Body).Decode(&createResp)
	resp.Body.Close()
	linkID := createResp.Data.ID

	// Redirect
	// Note: The router mounts "/" as fs, and "/{code}" for redirect.
	// Since we are using standard mux with wildcard in Go 1.22, /{code} matches anything not matched by others.
	// BUT, we have `mux.Handle("GET /", fs)`.
	// In Go 1.22, `GET /` matches everything ending in `/` or everything if no other match.
	// `GET /{code}` is more specific?
	// Actually, `GET /` matches request path `/`.
	// `GET /{code}` matches `/something`.
	// Let's test.

	redirectURL := tests.BuildRequestUrl(cfg.Server, "/"+prettyID)
	// We need to disable auto-redirect follow to check header
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, redirectURL, nil)
	resp, err = client.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	resp.Body.Close()

	if resp.StatusCode != http.StatusFound {
		t.Errorf("want 302, got %d", resp.StatusCode)
	}
	loc := resp.Header.Get("Location")
	if loc != originalURL {
		t.Errorf("want location %s, got %s", originalURL, loc)
	}

	// Check Analytics
	// Allow some time for async analytics to record
	time.Sleep(100 * time.Millisecond)

	analyticsURL := tests.BuildRequestUrl(cfg.Server, fmt.Sprintf("/api/links/%s/analytics", linkID))
	resp, err = tests.DoRequestWithCookie(ctx, http.MethodGet, analyticsURL, nil, cookie)
	if err != nil || resp.StatusCode != http.StatusOK {
		t.Fatalf("get analytics failed")
	}

	var analyticsResp struct {
		Stats struct {
			TotalClicks int64 `json:"total_clicks"`
		} `json:"stats"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&analyticsResp); err != nil {
		t.Fatalf("decode analytics failed: %v", err)
	}
	resp.Body.Close()

	if analyticsResp.Stats.TotalClicks != 1 {
		t.Errorf("want 1 click, got %d", analyticsResp.Stats.TotalClicks)
	}
}
