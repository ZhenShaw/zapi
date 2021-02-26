package zapi

import (
    "reflect"
    "sync"
)

// IHandler define then handler function type of api, it must be a func type.
// For example: func(*BaseContext)
type IHandler interface{}

type Api struct {
    Name string

    Methods []string

    Context IContext

    // Path if the path defined without prefix.
    Path string

    // final handler of api.
    Handler IHandler

    // fullPath contacts the prefix and path.
    fullPath string

    // handler chain, it stores the middleware and the final handler will be added to the end.
    handlers []IHandler

    // pool for reusing context.
    pool sync.Pool
}

func (api *Api) GetContext() IContext {

    if api.pool.New != nil {
        return api.pool.Get().(IContext)
    }
    api.pool.New = func() interface{} {
        return NewCtx(api.Context)
    }
    return api.pool.Get().(IContext)
}

func (api *Api) PutContext(ctx IContext) {
    ctx.Reset()
    api.pool.Put(ctx)
}

func NewCtx(c IContext) IContext {

    rv := reflect.ValueOf(c)
    rt := reflect.Indirect(rv).Type()

    value := reflect.New(rt)
    ctx := value.Interface().(IContext)

    return ctx
}
