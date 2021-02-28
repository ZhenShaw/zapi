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

func NewCtx(c IContext) IContext {

    rv := reflect.ValueOf(c)
    rt := reflect.Indirect(rv).Type()

    value := reflect.New(rt)
    ctx := value.Interface().(IContext)

    return ctx
}
