package zapi

import (
    "fmt"
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

    // fullPath contacts the prefix and path.
    fullPath string

    // handler chain, it stores the middleware and the final handler will be added to the end.
    Handlers []IHandler

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

func (api *Api) CheckHandlers() error {

    rt := reflect.TypeOf(api.Context)
    funcSign := fmt.Sprintf("func(%v)", rt)

    for _, h := range api.Handlers {
        sign := fmt.Sprint(reflect.TypeOf(h))
        if sign != funcSign && sign != iFaceFuncSign {
            err := fmt.Errorf("[%s] can only sign as: %s or %s, but got %s",
                api.fullPath, funcSign, iFaceFuncSign, sign)
            return err
        }
    }

    return nil
}
