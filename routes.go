package main

import (
	"fmt"
	"net/http"
)

func routes(mux *http.ServeMux, fs http.Handler) {
	mux.Handle("GET /health/", HandleHealth())
	mux.HandleFunc("GET /api/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println(r.URL.Path)
		w.Write([]byte("Cooking..."))
	})
	mux.Handle("GET /", fs)
}
