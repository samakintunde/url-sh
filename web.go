package main

import (
	"embed"
	"io/fs"
	"log/slog"
	"net/http"
)

//go:embed web/dist/*
var resources embed.FS

func InitWebServer() http.Handler {
	dist, err := fs.Sub(resources, "web/dist")
	if err != nil {
		slog.Error("couldn't open `web/dist` directory", err)
	}
	fs := http.FileServerFS(dist)
	return fs
}
