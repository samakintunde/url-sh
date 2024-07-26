package main

import (
	"context"
	"fmt"
	"net/http"
	db "url-shortener/db/sqlc"

	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
)

func routes(ctx context.Context, mux *http.ServeMux, fs http.Handler, queries *db.Queries, validate *validator.Validate, ut *ut.UniversalTranslator, trans ut.Translator, email Emailer) {
	mux.HandleFunc("GET /health/", HandleHealth())
	mux.Handle("POST /api/auth/signup", HandleSignup(ctx, queries, validate, ut, trans, email))
	mux.Handle("POST /api/auth/email-verification", HandleVerifyEmail(ctx, queries, validate, ut, trans))
	mux.Handle("POST /api/auth/login", HandleLogin(ctx, queries))
	mux.HandleFunc("GET /api/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println(r.URL.Path)
		w.Write([]byte("Cooking..."))
	})
	mux.Handle("GET /", fs)
}
