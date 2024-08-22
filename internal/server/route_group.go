package server

import (
	"net/http"
	"path"
)

type RouteGroup struct {
	prefix string
	mux    *http.ServeMux
}

func NewRouteGroup(prefix string, mux *http.ServeMux) *RouteGroup {
	return &RouteGroup{
		prefix: prefix,
		mux:    mux,
	}
}

func (rg *RouteGroup) Handle(pattern string, handler http.Handler) {
	fullPattern := path.Join(rg.prefix, pattern)
	rg.mux.Handle(fullPattern, handler)
}

func (rg *RouteGroup) HandleFunc(pattern string, handler func(http.ResponseWriter, *http.Request)) {
	fullPattern := path.Join(rg.prefix, pattern)
	rg.mux.HandleFunc(fullPattern, handler)
}
