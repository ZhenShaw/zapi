package zapi

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

// Router registers routes to be matched and dispatches a handler.
// It implements http.Handler interface, so it can be registered to serve.
// It extends the mux.Router struct, so it can most of features as mux.Router.
type Router struct {
	*mux.Router

	// a Router means a group of routes which has a same prefix.
	prefix string

	// store all apis defined.
	apis map[string]*Api

	// 含有相同前缀的api使用一系列中间件函数
	// Specifically, when prefix is empty or '/', it means using on all api.
	prefixHandlers map[string][]IHandler
}

func NewRouter() *Router {
	r := &Router{
		Router:         mux.NewRouter(),
		apis:           make(map[string]*Api),
		prefixHandlers: make(map[string][]IHandler),
	}
	return r
}

func (r *Router) Add(path string, context IContext, handler IHandler, options ...ApiOption) *mux.Route {
	api := &Api{
		Path:    path,
		Handler: handler,
		Context: context,
	}

	for _, fn := range options {
		fn.apply(api)
	}

	return r.AddApi(api)
}

func (r *Router) copy() *Router {
	return &Router{
		Router:         r.Router,
		apis:           r.apis,
		prefix:         r.prefix,
		prefixHandlers: r.prefixHandlers,
	}
}

func (r *Router) Sub(prefix string, middleware ...IHandler) *Router {
	router := r.copy()
	router.Router = router.PathPrefix(prefix).Subrouter()
	router.prefix = router.prefix + prefix
	router.Use(middleware...)
	return router
}

func (r *Router) AddApi(api *Api) *mux.Route {

	api.fullPath = r.prefix + api.Path

	if api.Name == "" {
		api.Name = uuid.New().String()
	}

	r.apis[api.Name] = api
	hf := func(w http.ResponseWriter, r *http.Request) {}
	route := r.Router.HandleFunc(api.Path, hf).Name(api.Name)

	return route
}

func (r *Router) Use(middleware ...IHandler) *Router {
	r.prefixHandlers[r.prefix] = append(r.prefixHandlers[r.prefix], middleware...)
	return r
}

// build handler chain for api.
// execute order: prefix handlers(middlewares) --> api handlers(middlewares) --> api final handler
func (r *Router) buildHandlerChain() {
	for _, api := range r.apis {
		for prefix, handlers := range r.prefixHandlers {
			if strings.HasPrefix(api.fullPath, prefix) {
				api.handlers = append(handlers, api.handlers...)
			}
		}
		api.handlers = append(api.handlers, api.Handler)
	}
}

// find methods of full-path the api supports.
func (r *Router) buildMethods() {
	for _, api := range r.apis {
		api.Methods = matchMethods(r, api.fullPath)
	}
}

func matchMethods(r *Router, fullPath string) []string {

	var AnyMethod = "NOT_SET"
	var methods = []string{AnyMethod, http.MethodGet, http.MethodHead, http.MethodPost,
		http.MethodPut, http.MethodPatch, http.MethodDelete,
		http.MethodConnect, http.MethodOptions, http.MethodTrace}

	var matched []string
	var match mux.RouteMatch
	var mockReq = &http.Request{}

	for _, method := range methods {

		mockReq.Method = method
		mockReq.URL, _ = url.Parse(fullPath)

		// if methods not set, it can also match value of mockReq.Method
		r.Match(mockReq, &match)

		if match.MatchErr != mux.ErrMethodMismatch {
			if method == AnyMethod {
				return []string{}
			}
			matched = append(matched, method)
		}
	}
	return matched
}

func (r *Router) Init() error {

	r.buildMethods()

	// check api definition
	var pathSet = make(map[string]struct{}, len(r.apis))
	var nameSet = make(map[string]struct{}, len(r.apis))
	for _, api := range r.apis {
		if api.Handler == nil {
			return fmt.Errorf("handler not set of api path: %s", api.fullPath)
		}
		if _, ok := pathSet[api.fullPath]; ok {
			return fmt.Errorf("duplicate api path: %s", api.fullPath)
		}
		if _, ok := nameSet[api.Name]; ok {
			return fmt.Errorf("duplicate api name: %s", api.Name)
		}
		pathSet[api.fullPath] = struct{}{}
	}

	r.buildHandlerChain()
	return nil
}
