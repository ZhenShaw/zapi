package zapi

import (
    "fmt"
    "github.com/gorilla/mux"
    "reflect"
    "sync"
)

// IHandler define then handler function type of api, it must be a func type.
// For example: func(*BaseContext)
type IHandler interface{}

type Api struct {
    name string

    // path if the path defined without prefix.
    path string

    // fullPath contacts the prefix and path.
    fullPath string

    methods []string

    context IContext

    // handler chain, it stores the middleware and the final handler will be added to the end.
    handlers []IHandler

    // pool for reusing context.
    pool sync.Pool

    route *mux.Route
}

func NewApi(path string, ctx IContext, handlers ...IHandler) *Api {
    return &Api{
        path:     path,
        context:  ctx,
        handlers: handlers,
    }
}

func (api *Api) Name(name string) *Api {
    api.name = name
    return api
}

func (api *Api) Methods(methods ...string) *Api {
    api.methods = methods
    if len(api.methods) > 0 && api.route != nil {
        api.route.Methods(api.methods...)
    }
    return api
}

func (api *Api) GetContext() IContext {

    if api.pool.New != nil {
        return api.pool.Get().(IContext)
    }
    api.pool.New = func() interface{} {
        return NewCtx(api.context)
    }
    return api.pool.Get().(IContext)
}

func (api *Api) PutContext(ctx IContext) {
    ctx.Reset()
    api.pool.Put(ctx)
}

func (api *Api) CheckHandlers() error {

    rt := reflect.TypeOf(api.context)
    funcSign := fmt.Sprintf("func(%v)", rt)

    for _, h := range api.handlers {
        sign := fmt.Sprint(reflect.TypeOf(h))
        if sign != funcSign && sign != iFaceFuncSign {
            err := fmt.Errorf("[%s] can only sign as: %s or %s, but got %s",
                api.fullPath, funcSign, iFaceFuncSign, sign)
            return err
        }
    }

    return nil
}
