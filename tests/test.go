package tests

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"time"
	"url-shortener/internal/config"
)

func WaitForReady(ctx context.Context, timeout time.Duration, endpoint string) error {
	client := http.Client{}
	start := time.Now()

	for {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)

		if err != nil {
			return fmt.Errorf("failed to create request: %w", err)
		}
		resp, err := client.Do(req)
		if err != nil {
			continue
		}
		if resp.StatusCode == http.StatusOK {
			slog.Info("Endpoint is ready!")
			resp.Body.Close()
			return nil
		}
		resp.Body.Close()

		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			if time.Since(start) >= timeout {
				return fmt.Errorf("timeout reached while waiting for endpoint")
			}
			time.Sleep(250 * time.Millisecond)
		}
	}
}

func DoRequest(ctx context.Context, method string, addr string, body io.Reader) (*http.Response, error) {
	return DoRequestWithCookie(ctx, method, addr, body, nil)
}

func DoRequestWithCookie(ctx context.Context, method string, addr string, body io.Reader, cookie *http.Cookie) (*http.Response, error) {
	client := http.Client{}
	req, err := http.NewRequestWithContext(ctx, method, addr, body)
	if err != nil {
		return nil, err
	}
	if cookie != nil {
		req.AddCookie(cookie)
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func BuildRequestUrl(serverCfg config.Server, path string) string {
	return fmt.Sprintf("http://127.0.0.1:%d%s", serverCfg.Port, path)
}

func BuildTestConfig(port int) config.Config {
	if port == 0 {
		port = 8100
	}
	// Use a temporary file for SQLite
	dbFile := filepath.Join(os.TempDir(), fmt.Sprintf("test_db_%d.sqlite", port))
	// Clean up previous run
	os.Remove(dbFile)

	return config.Config{
		Debug: true,
		Server: config.Server{
			Address:           fmt.Sprintf(":%d", port),
			Port:              port,
			TokenSymmetricKey: "9cb334a024c38493b9529a810969cc4e",
		},
		Log: config.Log{
			Level:  "debug",
			Format: "text",
		},
		Database: config.Database{
			Uri:                dbFile,
			MigrationSourceURL: "file://db/migrations",
		},
		SMTP: config.SMTP{
			Host:     "smtp.gmail.com",
			Port:     587,
			Username: "samakintunde37@gmail.com",
			Password: "password",
			From:     "test",
		},
	}
}
