package zapi

import (
    "fmt"
    "net/http"
    "net/url"
    "sort"
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

func (r *Router) NewApi(path string, context IContext, handlers ...IHandler) *Api {
    return NewApi(path, context, handlers...)
}

func (r *Router) Add(path string, context IContext, handlers ...IHandler) *Api {
    api := r.NewApi(path, context, handlers...)
    return r.addApi(api)
}

func (r *Router) Sub(prefix string, middleware ...IHandler) *Router {
    router := r.copy()
    router.Router = router.PathPrefix(prefix).Subrouter()
    router.prefix = router.prefix + prefix
    router.Use(middleware...)
    return router
}

func (r *Router) SubApi(prefix string, apis []*Api, middleware ...IHandler) *Router {
    router := r.Sub(prefix, middleware...)

    for _, api := range apis {
        router.addApi(api)
    }

    return router
}

func (r *Router) Use(middleware ...IHandler) *Router {
    r.prefixHandlers[r.prefix] = append(r.prefixHandlers[r.prefix], middleware...)
    return r
}

func (r *Router) addApi(api *Api) *Api {

    matchKey := uuid.New().String()
    emptyFun := func(w http.ResponseWriter, r *http.Request) {}

    api.route = r.Router.HandleFunc(api.path, emptyFun)
    api.Methods(api.methods...)
    api.fullPath = r.prefix + api.path
    api.route.Name(matchKey)

    r.apis[matchKey] = api

    return api
}

func (r *Router) copy() *Router {
    return &Router{
        Router:         r.Router,
        apis:           r.apis,
        prefix:         r.prefix,
        prefixHandlers: r.prefixHandlers,
    }
}

// build handler chain for api.
// execute order: prefix handlers(middlewares) + api handlers
func (r *Router) buildHandlerChain() {

    var prefixes []string
    for prefix, _ := range r.prefixHandlers {
        prefixes = append(prefixes, prefix)
    }
    sort.Strings(prefixes)

    for _, api := range r.apis {
        var mvs []IHandler
        for _, prefix := range prefixes {
            handlers := r.prefixHandlers[prefix]
            if strings.HasPrefix(api.fullPath, prefix) {
                mvs = append(mvs, handlers...)
            }
        }
        api.handlers = append(mvs, api.handlers...)
    }
}

// find methods of full-path the api supports.
func (r *Router) buildMethods() {
    for _, api := range r.apis {
        api.methods = matchMethods(r, api.fullPath)
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
    var pathMethodSet = make(map[string]struct{}, len(r.apis))
    for _, api := range r.apis {
        if len(api.handlers) == 0 {
            return fmt.Errorf("handlers not set of api path: %s", api.fullPath)
        }
        if _, ok := pathMethodSet[api.fullPath]; ok {
            return fmt.Errorf("duplicate api path: %s", api.fullPath)
        }
        pathMethodSet[api.fullPath] = struct{}{}
    }

    r.buildHandlerChain()

    for _, api := range r.apis {
        if err := api.CheckHandlers(); err != nil {
            return err
        }
    }

    return nil
}
