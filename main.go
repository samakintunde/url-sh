package main

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	db "url-shortener/db/sqlc"
	"url-shortener/internal/config"
	"url-shortener/internal/server"
	"url-shortener/internal/token"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite3"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	_ "github.com/mattn/go-sqlite3"
)

//go:embed db/migrations/*.sql
var migrationsFS embed.FS

func main() {
	ctx := context.Background()

	cfg, err := config.Load()

	if err != nil {
		fmt.Printf("Error loading config: %s\n", err)
		os.Exit(1)
		return
	}

	if err := run(ctx, cfg); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}

var interruptSignals = []os.Signal{
	os.Interrupt,
	syscall.SIGTERM,
	syscall.SIGINT,
}

func run(ctx context.Context, cfg config.Config) error {
	ctx, stop := signal.NotifyContext(ctx, interruptSignals...)

	// defer stop() // Don't defer stop here, it cancels context immediately when run returns if run is synchronous, but run blocks on select.
	// Wait, run() blocks on select.
	// If the test calls `go run(ctx, cfg)`, then `ctx` passed from test controls it.
	// `signal.NotifyContext` wraps the passed ctx.
	// If `stop` is called, `ctx` is canceled.
	// `defer stop()` means when `run` returns, stop is called. That is fine.
	// The problem might be how `signal.NotifyContext` interacts with `ctx.Done()` in the select.

	// If `ctx` passed from test is canceled, `ctx` (wrapped) is canceled.
	// `select` case `<-ctx.Done()` triggers.
	// Then `server.Shutdown` is called.
	// And `run` returns.

	// Issue in tests:
	// `go run(ctx, cfg)`
	// The test continues.
	// `tests.WaitForReady` checks health.

	// If `run` fails early (e.g. DB init), it returns error.
	// The goroutine exits.
	// `WaitForReady` keeps looping until timeout.

	// In the logs, we saw:
	// 2026/01/09 10:45:18 INFO Run mode: Debug=true
	// 2026/01/09 10:45:18 INFO initialised DB
	// 2026/01/09 10:45:18 INFO migrations completed successfully
	// 2026/01/09 10:45:18 INFO starting server on 0.0.0.0:8101
	// ... then timeout.

	// So server started.
	// Why is it not reachable?
	// `tests.WaitForReady` tries `http://127.0.0.1:8101/health`.
	// Logs show `starting server on 0.0.0.0:8101`.
	// 0.0.0.0 should be reachable via 127.0.0.1.

	// Wait, `tests.WaitForReady` creates a request with context.
	// `req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)`
	// `ctx` passed to `WaitForReady` is the same `ctx` passed to `run`.
	// NO. In test:
	// `ctx := context.Background()`
	// `ctx, cancel := context.WithCancel(ctx)`
	// `go run(ctx, cfg)`
	// `tests.WaitForReady(ctx, ...)`

	// If `run` modifies `ctx`? No, it shadows it: `ctx, stop := signal.NotifyContext(ctx, ...)`
	// But `tests.WaitForReady` uses the original `ctx`.

	// Is it possible `fs := InitWebServer()` is taking too long or failing silently?
	// `InitWebServer` in `web.go`?
	// Let's check `web.go`.

	defer stop()

	slog.Info("Run mode:", "Debug", cfg.Debug)

	sqliteDB, err := initDB(cfg.Database)

	if err != nil {
		slog.Error("initDB", "error", err)
		return err
	}

	defer sqliteDB.Close()

	slog.Info("initialised DB")

	err = runMigration(sqliteDB)

	if err != nil {
		slog.Error("runMigration", "error", err)
		return err
	}

	queries := db.New(sqliteDB)

	fs := InitWebServer()

	tokenMaker, err := token.NewPasetoMaker(cfg.Server.TokenSymmetricKey)

	if err != nil {
		slog.Error("token.NewPasetoMaker", "error", err)
		return err
	}

	srv := server.New(ctx, cfg, fs, queries, tokenMaker)

	httpServer := &http.Server{
		Addr:         cfg.Server.Address,
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
			slog.Error("httpServer.ListenAndServe", "error", err)
			return err
		}
	case <-ctx.Done():
		const timeout = 1 * time.Second
		shutdownCtx, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()
		if err := httpServer.Shutdown(shutdownCtx); err != nil {
			slog.Error(fmt.Sprintf("server failed to shut down gracefully in %v", timeout), "error", err)
			if err := httpServer.Close(); err != nil {
				slog.Error("httpServer.Close", "error", err)
				return err
			}
		}
		slog.Info("server shut down")
	}

	return nil
}

func initDB(cfg config.Database) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", cfg.Uri)
	if err != nil {
		return nil, err
	}
	if err = db.Ping(); err != nil {
		return nil, err
	}
	return db, nil
}

func runMigration(db *sql.DB) error {
	// Will wrap each migration in an implicit transaction by default
	driver, err := sqlite3.WithInstance(db, &sqlite3.Config{})
	if err != nil {
		return fmt.Errorf("failed to create sqlite3 driver: %w", err)
	}

	d, err := iofs.New(migrationsFS, "db/migrations")
	if err != nil {
		return fmt.Errorf("Could not create ioFS driver: %v", err)
	}

	migration, err := migrate.NewWithInstance("iofs", d, "sqlite3", driver)

	if err != nil {
		return fmt.Errorf("failed to create migration instance: %w", err)
	}

	err = migration.Up()

	if err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	slog.Info("migrations completed successfully")

	return nil
}
