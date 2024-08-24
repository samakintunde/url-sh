package server

import (
	"net/http"
	"path"
	"strings"
)

type RouteGroup struct {
	prefix      string
	mux         *http.ServeMux
	middlewares []func(http.Handler) http.Handler
}

func NewRouteGroup(prefix string, mux *http.ServeMux) *RouteGroup {
	if strings.Contains(prefix, " ") {
		panic("route group prefix cannot contain spaces. You are probably trying to use a pattern `POST /route` which is not supported on route groups")
	}
	return &RouteGroup{
		prefix: prefix,
		mux:    mux,
	}
}

func (rg *RouteGroup) Group(prefix string) *RouteGroup {
	newPrefix := joinPattern(rg.prefix, prefix)
	subrg := NewRouteGroup(newPrefix, rg.mux)
	return subrg
}

func (rg *RouteGroup) Use(middleware func(next http.Handler) http.Handler) {
	rg.middlewares = append(rg.middlewares, middleware)
}

func (rg *RouteGroup) Handle(pattern string, handler http.Handler) {
	fullPattern := joinPattern(rg.prefix, pattern)
	wrappedHandler := handler
	for i := len(rg.middlewares) - 1; i >= 0; i-- {
		wrappedHandler = rg.middlewares[i](wrappedHandler)
	}
	rg.mux.Handle(fullPattern, wrappedHandler)
}

func (rg *RouteGroup) HandleFunc(pattern string, handler func(http.ResponseWriter, *http.Request)) {
	fullPattern := joinPattern(rg.prefix, pattern)
	// Feels very hacky. Might just remove support for the HandleFunc altogether
	wrappedHandler := http.Handler(http.HandlerFunc(handler))
	for i := len(rg.middlewares) - 1; i >= 0; i-- {
		wrappedHandler = rg.middlewares[i](wrappedHandler)
	}
	rg.mux.HandleFunc(fullPattern, wrappedHandler.ServeHTTP)
}

func joinPattern(prefix string, pattern string) string {
	var fullPattern string

	// Handles cases where `POST /route` is used
	if strings.Contains(pattern, " ") {
		splitPattern := strings.Split(pattern, " ")

		if len(splitPattern) == 2 {
			fullPattern = splitPattern[0] + " " + path.Join(prefix, splitPattern[1])
		} else {
			fullPattern = path.Join(prefix, splitPattern[0])
		}

		return fullPattern
	}

	return path.Join(prefix, pattern)
}
