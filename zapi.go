package zapi

import (
    "net/http"

    "github.com/gorilla/mux"
    "github.com/zhenshaw/go-lib/logs"
)

const defaultAddr = ":8080"

func debugLog(f interface{}, v ...interface{}) {
    logs.Debug(f, v...)
}

type app struct {
    Router *Router
    Addr   string
}

func NewApp() *app {
    return &app{
        Router: NewRouter(),
        Addr:   defaultAddr,
    }
}

func (app *app) initApp() error {
    if err := app.Router.Init(); err != nil {
        return err
    }
    return nil
}

func (app *app) GetRouter() *Router {
    return app.Router
}

func (app *app) Run(addr ...string) error {

    if len(addr) != 0 {
        app.Addr = addr[0]
    }

    if err := app.initApp(); err != nil {
        return err
    }

    //todo: graceful shutdown
    debugLog("HTTP server listening on", app.Addr)
    return http.ListenAndServe(app.Addr, app)
}

func (app *app) RunLTS(certFile, keyFile string) error {

    if err := app.initApp(); err != nil {
        return err
    }

    debugLog("HTTP server listening on", app.Addr)
    return http.ListenAndServeTLS(app.Addr, certFile, keyFile, app)
}

func (app *app) ServeHTTP(w http.ResponseWriter, r *http.Request) {

    var match mux.RouteMatch
    var handler http.Handler
    if app.Router.Match(r, &match) {
        handler = match.Handler
        r = requestWithVars(r, match.Vars)
        r = requestWithRoute(r, match.Route)
    }

    if handler == nil && match.MatchErr == mux.ErrMethodMismatch {
        w.WriteHeader(http.StatusMethodNotAllowed)
        return
    }

    if handler == nil {
        http.NotFoundHandler().ServeHTTP(w, r)
        return
    }

    api, ok := app.Router.apis[match.Route.GetName()]
    if !ok {
        http.NotFoundHandler().ServeHTTP(w, r)
        return
    }

    app.HandleRequest(api, w, r)
}

func (app *app) HandleRequest(api *Api, w http.ResponseWriter, r *http.Request) {
    ctx := api.GetContext()
    ctx.Init(api.handlers, w, r)
    ctx.Start(ctx)
    api.PutContext(ctx)
}
