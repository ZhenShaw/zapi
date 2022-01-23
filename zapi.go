package zapi

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/zhenshaw/go-lib/logs"
	"google.golang.org/grpc"
)

const defaultAddr = ":8080"

func debugLog(f interface{}, v ...interface{}) {
	logs.Debug(f, v...)
}

type app struct {
	server *http.Server
	Router *Router
	Addr   string
	GRPC   *grpc.Server
	Cors   bool
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

	debugLog("HTTP server listening on", app.Addr)
	app.server = &http.Server{Addr: app.Addr, Handler: http.HandlerFunc(app.handle)}
	return app.server.ListenAndServe()
}

func (app *app) RunLTS(certFile, keyFile string) error {

	if err := app.initApp(); err != nil {
		return err
	}

	debugLog("HTTPS server listening on", app.Addr)

	app.server = &http.Server{Addr: app.Addr, Handler: http.HandlerFunc(app.handle)}
	return app.server.ListenAndServeTLS(certFile, keyFile)
}

func (app *app) handle(w http.ResponseWriter, r *http.Request) {
	if app.GRPC != nil && r.ProtoMajor == 2 && strings.HasPrefix(
		r.Header.Get("Content-Type"), "application/grpc") {
		app.GRPC.ServeHTTP(w, r)
	} else {
		app.ServeHTTP(w, r)
	}
}

func (app *app) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	if app.Cors {
		w.Header().Add("Access-Control-Allow-Headers", "*")
		w.Header().Add("Access-Control-Allow-Origin", r.Header.Get("Origin"))
		w.Header().Add("Access-Control-Allow-Credentials", "true")
		w.Header().Add("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PATCH, PUT, DELETE")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
	}

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
	ctx.Finish()
	api.PutContext(ctx)
}

func (app *app) Shutdown() {
	ctx, reset := signal.NotifyContext(context.Background(), os.Interrupt)
	<-ctx.Done()
	reset()

	timeoutCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	debugLog("shutting down gracefully, press Ctrl+C again to force exit")
	if err := app.server.Shutdown(timeoutCtx); err != nil {
		debugLog("graceful shutdown fail:", err)
		os.Exit(1)
	}
}
