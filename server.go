package main

import (
	"net/http"
)

func NewServer(fs http.Handler) http.Handler {
	mux := http.NewServeMux()
	routes(mux, fs)
	return mux
}
