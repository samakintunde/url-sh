package config

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/spf13/viper"
)

type Config struct {
	Log      Log
	SMTP     SMTP
	Database Database
	Server   Server
	Debug    bool
}

type EnvLoader interface {
	Load() (Config, error)
}

func Load() (Config, error) {
	v := viper.New()

	var envFile string

	if os.Getenv("DEBUG") == "true" {
		envFile = ".env.local"
	}

	v.SetConfigFile(envFile)
	v.SetConfigType("env")

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			slog.Warn("couldn't find config file", "file", envFile, "error", err)
			return Config{}, err
		}
		slog.Info("Config file not found. Using environment variables.")
	}

	v.AutomaticEnv()

	config := initConfig(v)

	return config, nil
}

func initConfig(v *viper.Viper) Config {
	debug := v.GetBool("DEBUG")

	port := v.GetInt("PORT")
	serverConfig := Server{
		Address:           fmt.Sprintf("0.0.0.0:%d", port),
		Port:              port,
		TokenSymmetricKey: v.GetString("TOKEN_SYMMETRIC_KEY"),
	}

	var logLevel string
	var logFormat string
	if debug {
		logLevel = "debug"
		logFormat = "text"
	} else {
		logFormat = "json"
	}

	logConfig := Log{
		Level:  logLevel,
		Format: logFormat,
	}

	databaseConfig := Database{
		Uri: v.GetString("DATABASE_URI"),
	}

	smtpConfig := SMTP{
		Host:     v.GetString("SMTP_HOST"),
		Port:     v.GetInt("SMTP_PORT"),
		Username: v.GetString("SMTP_USERNAME"),
		Password: v.GetString("SMTP_PASSWORD"),
		From:     v.GetString("SMTP_EMAIL"),
	}

	return Config{
		Debug:    debug,
		Log:      logConfig,
		SMTP:     smtpConfig,
		Database: databaseConfig,
		Server:   serverConfig,
	}
}
