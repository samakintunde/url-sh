package tests

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
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
	client := http.Client{}
	req, err := http.NewRequestWithContext(ctx, method, addr, body)
	if err != nil {
		return nil, err
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

func BuildTestConfig() config.Config {
	port := 8100
	return config.Config{
		Debug: true,
		Server: config.Server{
			Address:           fmt.Sprintf("0.0.0.0:%d", port),
			Port:              port,
			TokenSymmetricKey: "bB34U3baPLuXWmBsol15g0aeV5VxF43f",
		},
		Log: config.Log{
			Level:  "debug",
			Format: "text",
		},
		Database: config.Database{
			Uri:                ":memory:",
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
