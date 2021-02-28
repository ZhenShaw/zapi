// using go:linkname to force export private function.
package zapi

import (
    "net/http"
    _ "unsafe"

    "github.com/gorilla/mux"
)

//go:linkname requestWithVars github.com/gorilla/mux.requestWithVars
func requestWithVars(r *http.Request, vars map[string]string) *http.Request

//go:linkname requestWithRoute github.com/gorilla/mux.requestWithRoute
func requestWithRoute(r *http.Request, route *mux.Route) *http.Request
