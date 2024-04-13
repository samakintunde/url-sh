package main

import "fmt"

type Config struct {
	HttpAddr string
	Port     string
}

func InitConfig(env func(string, string) string) Config {
	port := env("PORT", "8080")

	return Config{
		HttpAddr: fmt.Sprintf("0.0.0.0:%s", port),
		Port:     port,
	}
}
