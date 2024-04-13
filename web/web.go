package web

import (
	"embed"
	"io/fs"
	"log/slog"
	"net/http"
)

//go:embed dist/*
var resources embed.FS

func InitWebServer() http.Handler {
	dist, err := fs.Sub(resources, "dist")
	if err != nil {
		slog.Error("couldn't open `dist` directory", err)
	}
	fs := http.FileServerFS(dist)
	return fs
}
