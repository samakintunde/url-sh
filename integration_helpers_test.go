package main

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"testing"
	"url-shortener/internal/config"
	"url-shortener/tests"

	_ "github.com/mattn/go-sqlite3"
)

// Helper to get verification code from DB
func getVerificationCode(dbPath, email string) (string, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return "", err
	}
	defer db.Close()

	var code string
	err = db.QueryRow("SELECT code FROM email_verifications WHERE email = ? ORDER BY created_at DESC LIMIT 1", email).Scan(&code)
	return code, err
}

func createUserAndLogin(t *testing.T, ctx context.Context, cfg config.Config, email, password string) *http.Cookie {
	// 1. Signup
	signupBody := fmt.Sprintf(`{"email": "%s","password": "%s","first_name": "Test","last_name": "User"}`, email, password)
	resp, err := tests.DoRequest(ctx, http.MethodPost, tests.BuildRequestUrl(cfg.Server, "/api/auth/signup"), bytes.NewReader([]byte(signupBody)))
	if err != nil {
		t.Fatalf("signup failed: %v", err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("signup returned %d", resp.StatusCode)
	}

	// 2. Verify Email
	code, err := getVerificationCode(cfg.Database.Uri, email)
	if err != nil {
		t.Fatalf("failed to get verification code: %v", err)
	}

	verifyBody := fmt.Sprintf(`{"code": "%s"}`, code)
	resp, err = tests.DoRequest(ctx, http.MethodPost, tests.BuildRequestUrl(cfg.Server, "/api/auth/email-verification"), bytes.NewReader([]byte(verifyBody)))
	if err != nil {
		t.Fatalf("verify failed: %v", err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("verify returned %d", resp.StatusCode)
	}

	// 3. Login
	loginBody := fmt.Sprintf(`{"email": "%s","password": "%s"}`, email, password)
	resp, err = tests.DoRequest(ctx, http.MethodPost, tests.BuildRequestUrl(cfg.Server, "/api/auth/login"), bytes.NewReader([]byte(loginBody)))
	if err != nil {
		t.Fatalf("login failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("login returned %d", resp.StatusCode)
	}

	// Extract cookie
	cookies := resp.Cookies()
	for _, c := range cookies {
		if c.Name == "access_token" {
			return c
		}
	}
	t.Fatal("access_token cookie not found")
	return nil
}
