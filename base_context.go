package zapi

import (
    "net/http"
    "reflect"
)

const iFaceFuncSign = "func(zapi.IContext)"

type IContext interface {
    Init([]IHandler, http.ResponseWriter, *http.Request)
    Start(IContext)

    Call(IHandler)
    Next()
    Finish()
    Reset()

    GetRequest() *http.Request
    GetWriter() http.ResponseWriter
}

func NewCtx(c IContext) IContext {

    rv := reflect.ValueOf(c)
    rt := reflect.Indirect(rv).Type()

    value := reflect.New(rt)
    ctx := value.Interface().(IContext)

    return ctx
}

type BaseContext struct {
    Request *http.Request
    Writer  http.ResponseWriter

    handlers []IHandler

    // mark index in handlers chain
    index int

    ctx IContext
}

// Init should init all field though it may zero val.
func (z *BaseContext) Init(handlers []IHandler, w http.ResponseWriter, r *http.Request) {
    z.Writer = w
    z.Request = r
    z.handlers = handlers
    z.index = 0
    z.ctx = nil
}

/*
Call assert and call the handler.
If implement BaseContext in a new struct, it recommend to rewrite this function,
for example:

func (c *MyCtx) Call(handler IHandler) {
	switch handleFun := handler.(type) {
	case func(*MyCtx):
		handleFun(c)
	default:
		c.BaseContext.Call(handler)
	}
}
*/
func (z *BaseContext) Call(handler IHandler) {

    switch handleFun := handler.(type) {
    case func(*BaseContext):
        handleFun(z)
    case func(IContext):
        handleFun(z.ctx)
    default:
        fn := reflect.ValueOf(handler)
        fn.Call([]reflect.Value{reflect.ValueOf(z.ctx)})
    }
}

// Start execute the first handler of handler chain.
// Warning: Do not rewrite unless you known what you are doing.
func (z *BaseContext) Start(ctx IContext) {
    z.ctx = ctx
    if len(z.handlers) != 0 && z.index == 0 {
        z.index++
        z.ctx.Call(z.handlers[0])
    }
}

// Next execute the next handler of handler chain which determined by increasing index.
// It should call in every middleware manually when all go through.
// Warning: Do not rewrite unless you known what you are doing.
func (z *BaseContext) Next() {
    if len(z.handlers) <= z.index {
        return
    }
    handler := z.handlers[z.index]
    z.index++
    z.ctx.Call(handler)

    if len(z.handlers) == z.index {
        z.ctx.Finish()
    }
}

// Finish allow you do some extra work after the last handler was executed.
func (z *BaseContext) Finish() {}

// Reset set zero value for reusing context struct in sync pool.
func (z *BaseContext) Reset() {
    z.Request = nil
    z.Writer = nil
    z.handlers = nil
    z.index = 0
    z.ctx = nil
}

// GetRequest returns *http.Request
func (z *BaseContext) GetRequest() *http.Request {
    return z.Request
}

// GetWriter returns http.ResponseWriter
func (z *BaseContext) GetWriter() http.ResponseWriter {
    return z.Writer
}
