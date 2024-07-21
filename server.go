package main

import (
	"context"
	"net/http"
	db "url-shortener/db/sqlc"

	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
)

func NewServer(ctx context.Context, fs http.Handler, queries *db.Queries, validate *validator.Validate, ut *ut.UniversalTranslator, trans ut.Translator) http.Handler {
	mux := http.NewServeMux()
	routes(ctx, mux, fs, queries, validate, ut, trans)
	return mux
}
