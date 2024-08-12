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
	"url-shortener/web"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite3"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/mattn/go-sqlite3"
)

func main() {
	ctx := context.Background()

	if err := run(ctx, env); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}

var interruptSignals = []os.Signal{
	os.Interrupt,
	syscall.SIGTERM,
	syscall.SIGINT,
}

func run(ctx context.Context, env func(string, string) string) error {
	ctx, stop := signal.NotifyContext(ctx, interruptSignals...)

	defer stop()

	config := InitConfig(env)
	sqliteDB, err := initDB(config)

	if err != nil {
		slog.Error("couldn't init db", err)
		return err
	}

	defer sqliteDB.Close()

	err = runMigration(sqliteDB, config)

	if err != nil {
		slog.Error("couldn't run migrations", err)
		return err
	}

	_ = db.New(sqliteDB)

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

func initDB(cfg Config) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", cfg.DatabaseUri)
	if err != nil {
		return nil, err
	}
	if err = db.Ping(); err != nil {
		return nil, err
	}
	return db, nil
}

func runMigration(db *sql.DB, config Config) error {
	// Will wrap each migration in an implicit transaction by default
	driver, err := sqlite3.WithInstance(db, &sqlite3.Config{})
	if err != nil {
		return err
	}

	migration, err := migrate.NewWithDatabaseInstance(config.MigrationSourceURL, "sqlite3", driver)

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
