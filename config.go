package main

import "fmt"

type Config struct {
	HttpAddr           string
	Port               string
	Environment        string
	DatabaseUri        string
	MigrationSourceURL string
}

func InitConfig(env func(string, string) string) Config {
	port := env("PORT", "8080")
	environment := env("ENVIRONMENT", "debug")
	databaseUri := env("DATABASE_URI", "./database.db")
	migrationSourceURL := env("MIGRATION_SOURCE_URL", "file://db/migrations")

	return Config{
		HttpAddr:           fmt.Sprintf("0.0.0.0:%s", port),
		Port:               port,
		Environment:        environment,
		DatabaseUri:        databaseUri,
		MigrationSourceURL: migrationSourceURL,
	}
}
