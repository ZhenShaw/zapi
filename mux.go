// code in this file is copied from github.com/gorilla/mux/mux.go
// for using in zapi directly.

package zapi

import (
	"context"
	"net/http"

	"github.com/gorilla/mux"
)

type contextKey int

const (
	varsKey contextKey = iota
	routeKey
)

func requestWithVars(r *http.Request, vars map[string]string) *http.Request {
	ctx := context.WithValue(r.Context(), varsKey, vars)
	return r.WithContext(ctx)
}

func requestWithRoute(r *http.Request, route *mux.Route) *http.Request {
	ctx := context.WithValue(r.Context(), routeKey, route)
	return r.WithContext(ctx)
}
