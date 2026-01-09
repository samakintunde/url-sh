package main

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"
	"url-shortener/tests"
)

func TestUserProfile(t *testing.T) {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	t.Cleanup(cancel)

	cfg := tests.BuildTestConfig(8103)
	go run(ctx, cfg)

	if err := tests.WaitForReady(ctx, 10*time.Second, fmt.Sprintf("http://localhost:%d/health", cfg.Server.Port)); err != nil {
		t.Fatalf("server not ready: %v", err)
	}

	email := "profile@example.com"
	password := "SecurePass123."
	cookie := createUserAndLogin(t, ctx, cfg, email, password)

	// Get Profile
	resp, err := tests.DoRequestWithCookie(ctx, http.MethodGet, tests.BuildRequestUrl(cfg.Server, "/api/user/me"), nil, cookie)
	if err != nil || resp.StatusCode != http.StatusOK {
		t.Fatalf("get profile failed")
	}
	resp.Body.Close()

	// Change Password
	newPass := "NewSecurePass456."
	changePassBody := fmt.Sprintf(`{"old_password": "%s", "new_password": "%s"}`, password, newPass)
	resp, err = tests.DoRequestWithCookie(ctx, http.MethodPost, tests.BuildRequestUrl(cfg.Server, "/api/user/change-password"), bytes.NewReader([]byte(changePassBody)), cookie)
	if err != nil || resp.StatusCode != http.StatusOK {
		t.Fatalf("change password failed: %d", resp.StatusCode)
	}
	resp.Body.Close()

	// Verify login with new password
	loginBody := fmt.Sprintf(`{"email": "%s","password": "%s"}`, email, newPass)
	resp, err = tests.DoRequest(ctx, http.MethodPost, tests.BuildRequestUrl(cfg.Server, "/api/auth/login"), bytes.NewReader([]byte(loginBody)))
	if err != nil || resp.StatusCode != http.StatusOK {
		t.Fatalf("login with new password failed")
	}
	resp.Body.Close()
}
