package zapi

import (
    "net/http"
    "reflect"
)

func Wrap(handleFun func(http.ResponseWriter, *http.Request)) IHandler {
    return func(ctx IContext) {
        handleFun(ctx.GetWriter(), ctx.GetRequest())
    }
}

func WrapHttp(handler IHandler) IHandler {
    switch h := handler.(type) {
    case func(http.ResponseWriter, *http.Request):
        return Wrap(h)
    case http.Handler:
        return Wrap(h.ServeHTTP)
    default:
        return handler
    }
}

func NewCtx(c IContext) IContext {

    rv := reflect.ValueOf(c)
    rt := reflect.Indirect(rv).Type()

    value := reflect.New(rt)
    ctx := value.Interface().(IContext)

    return ctx
}
