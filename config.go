package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/spf13/viper"
)

type SMTPConfig struct {
	Host     string
	Port     int
	Username string
	Password string
	From     string
}

type Config struct {
	HttpAddr           string
	Port               int
	Debug              bool
	DatabaseUri        string
	MigrationSourceURL string
	SMTP               SMTPConfig
}

type EnvLoader interface {
	Load() (Config, error)
}

func LoadEnvFile() (Config, error) {
	v := viper.New()

	var envFile string

	if os.Getenv("DEBUG") != "true" {
		envFile = ".env.production"
	} else {
		envFile = ".env.local"
	}

	v.SetConfigFile(envFile)
	v.SetConfigType("env")

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			slog.Warn("couldn't find config file", "file", envFile, "error", err)
			return Config{}, err
		}
		slog.Info("Config file not found. Using environment variables.", "file", envFile)
	}

	v.AutomaticEnv()

	config := InitConfig(v)

	return config, nil
}

func InitConfig(v *viper.Viper) Config {
	debug := v.GetBool("DEBUG")
	port := v.GetInt("PORT")
	databaseUri := v.GetString("DATABASE_URI")
	migrationSourceURL := v.GetString("MIGRATION_SOURCE_URL")
	smtpHost := v.GetString("SMTP_HOST")
	smtpPort := v.GetInt("SMTP_PORT")
	smtpUsername := v.GetString("SMTP_USERNAME")
	smtpPassword := v.GetString("SMTP_PASSWORD")
	smtpFrom := v.GetString("SMTP_EMAIL")

	return Config{
		HttpAddr:           fmt.Sprintf("0.0.0.0:%d", port),
		Port:               port,
		Debug:              debug,
		DatabaseUri:        databaseUri,
		MigrationSourceURL: migrationSourceURL,
		SMTP: SMTPConfig{
			Host:     smtpHost,
			Port:     smtpPort,
			Username: smtpUsername,
			Password: smtpPassword,
			From:     smtpFrom,
		},
	}
}
