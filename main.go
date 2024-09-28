package main

import (
	"context"
	"database/sql"
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
	"url-shortener/web"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite3"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/mattn/go-sqlite3"
)

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

	defer stop()

	slog.Info("Run mode:", "Debug", cfg.Debug)

	sqliteDB, err := initDB(cfg.Database)

	if err != nil {
		slog.Error("couldn't init db", err)
		return err
	}

	defer sqliteDB.Close()

	err = runMigration(sqliteDB, cfg.Database)

	if err != nil {
		slog.Error("couldn't run migrations", err)
		return err
	}

	queries := db.New(sqliteDB)

	fs := web.InitWebServer()

	tokenMaker, err := token.NewPasetoMaker(cfg.Server.TokenSymmetricKey)

	if err != nil {
		slog.Error("couldn't create token maker", err)
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
			slog.Error("error listening and serving", err)
		}
	case <-ctx.Done():
		const timeout = 1 * time.Second
		shutdownCtx, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()
		if err := httpServer.Shutdown(shutdownCtx); err != nil {
			slog.Error(fmt.Sprintf("server failed to shut down gracefully in %v", timeout), err)
			if err := httpServer.Close(); err != nil {
				slog.Error("closed server immediately", err)
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

func runMigration(db *sql.DB, cfg config.Database) error {
	// Will wrap each migration in an implicit transaction by default
	driver, err := sqlite3.WithInstance(db, &sqlite3.Config{})
	if err != nil {
		return err
	}

	migration, err := migrate.NewWithDatabaseInstance(cfg.MigrationSourceURL, "sqlite3", driver)

	if err != nil {
		return err
	}

	err = migration.Up()

	if err != nil && err != migrate.ErrNoChange {
		return err
	}

	slog.Info("migrations completed successfully")

	return nil
}
