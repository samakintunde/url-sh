package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"testing"
	"time"
	"url-shortener/tests"
)

func TestUserSignup(t *testing.T) {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)

	t.Cleanup(cancel)

	cfg := tests.BuildTestConfig()

	go run(ctx, cfg)

	timeout := 5 * time.Second
	err := tests.WaitForReady(ctx, timeout, fmt.Sprintf("http://127.0.0.1:%d/health", cfg.Server.Port))

	if err != nil {
		log.Fatalf("couldn't start the server in %fs", timeout.Seconds())
	}

	t.Run("it should return 400 for bad requests", func(t *testing.T) {
		cases := []io.Reader{
			bytes.NewReader([]byte("")),
			bytes.NewReader([]byte("{}")),
			bytes.NewReader([]byte("{\"email\": \"user\"}")),
		}
		addr := tests.BuildRequestUrl(cfg.Server, "/api/auth/signup")

		for _, body := range cases {
			resp, err := tests.DoRequest(ctx, http.MethodPost, addr, body)
			if err != nil {
				t.Fatalf("failed: %v", err)
			}
			defer resp.Body.Close()
			if resp.StatusCode != http.StatusBadRequest {
				t.Fatalf("want: %d, got: %d", http.StatusBadRequest, resp.StatusCode)
			}
		}
	})

	t.Run("it should return 201 for good requests", func(t *testing.T) {
		cases := []io.Reader{
			bytes.NewReader([]byte("{\"email\": \"user1@example.com\",\"password\": \"PBTsVser1.\",\"first_name\": \"John\",\"last_name\": \"Doe\"}")),
			bytes.NewReader([]byte("{\"email\": \"user3@example.com\",\"password\": \"PBTsVser2.\",\"first_name\": \"John\",\"last_name\": \"Doe\"}")),
			bytes.NewReader([]byte("{\"email\": \"user2@example.com\",\"password\": \"PBTsVser3.\",\"first_name\": \"John\",\"last_name\": \"Doe\"}")),
		}
		addr := tests.BuildRequestUrl(cfg.Server, "/api/auth/signup")

		for _, body := range cases {
			resp, err := tests.DoRequest(ctx, http.MethodPost, addr, body)
			if err != nil {
				t.Fatalf("failed: %v", err)
			}
			defer resp.Body.Close()
			if resp.StatusCode != http.StatusCreated {
				t.Fatalf("want: %d, got: %d", http.StatusCreated, resp.StatusCode)
			}
		}
	})
}
