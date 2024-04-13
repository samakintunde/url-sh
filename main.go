package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	"url-shortener/web"
)

func main() {
	ctx := context.Background()

	if err := run(ctx, env); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, env func(string, string) string) error {
	ctx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)

	defer stop()

	config := InitConfig(env)
	fs := web.InitWebServer()
	srv := NewServer(fs)

	httpServer := &http.Server{
		Addr:         config.HttpAddr,
		Handler:      srv,
		ReadTimeout:  2 * time.Second,
		WriteTimeout: 2 * time.Second,
		IdleTimeout:  2 * time.Second,
	}

	serverErr := make(chan error, 1)

	go func() {
		slog.Info(fmt.Sprintf("starting server on %s", httpServer.Addr))
		serverErr <- httpServer.ListenAndServe()
	}()

	select {
	case err := <-serverErr:
		if err != nil && err != http.ErrServerClosed {
			slog.Error("error listening and serving", err)
		}
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(ctx, 1*time.Second)
		defer cancel()
		if err := httpServer.Shutdown(shutdownCtx); err != nil {
			slog.Error("error shutting down server gracefully", err)
		}
		slog.Info("server shut down")
	}

	return nil
}
