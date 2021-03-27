package zapi

import (
    "net/http"
    "reflect"
)

type IContext interface {
    Init([]IHandler, http.ResponseWriter, *http.Request)
    Start(IContext)
    Prepare() bool
    Call(IHandler)
    Next()
    Finish()
    Reset()
    GetRequest() *http.Request
    GetWriter() http.ResponseWriter
}

type baseContext struct {
    Request *http.Request
    Writer  http.ResponseWriter

    handlers []IHandler

    // mark index in handlers chain
    index int

    ctx IContext
}

// Init should init all field though it may zero val.
func (z *baseContext) Init(handlers []IHandler, w http.ResponseWriter, r *http.Request) {
    z.Writer = w
    z.Request = r
    z.handlers = handlers
    z.index = 0
    z.ctx = nil
}

/*
Call assert and call the handler.
If implement baseContext in a new struct, it recommend to rewrite this function,
for example:

func (c *MyCtx) Call(handler IHandler) {
	switch handleFun := handler.(type) {
	case func(*MyCtx):
		handleFun(c)
	default:
		c.baseContext.Call(handler)
	}
}
*/
func (z *baseContext) Call(handler IHandler) {

    switch handleFun := handler.(type) {
    case func(*baseContext):
        handleFun(z)
    case func(IContext):
        handleFun(z.ctx)
    default:
        fn := reflect.ValueOf(handler)
        fn.Call([]reflect.Value{reflect.ValueOf(z.ctx)})
    }
}

func (z *baseContext) Prepare() bool { return true }

// Start execute the first handler of handler chain.
// Warning: Do not rewrite unless you known what you are doing.
func (z *baseContext) Start(ctx IContext) {
    z.ctx = ctx
    if len(z.handlers) != 0 && z.index == 0 {
        z.index++
        z.ctx.Call(z.handlers[0])
    }
}

// Next execute the next handler of handler chain which determined by increasing index.
// It should call in every middleware manually when all go through.
// Warning: Do not rewrite unless you known what you are doing.
// Next only execute if not Responded.
func (z *baseContext) Next() {
    if len(z.handlers) <= z.index {
        return
    }
    handler := z.handlers[z.index]
    z.index++
    z.ctx.Call(handler)
}

// Finish allow you do some extra work after the last handler was executed.
func (z *baseContext) Finish() {}

// Reset set zero value for reusing context struct in sync pool.
func (z *baseContext) Reset() {
    z.Request = nil
    z.Writer = nil
    z.handlers = nil
    z.index = 0
    z.ctx = nil
}

// GetRequest returns *http.Request
func (z *baseContext) GetRequest() *http.Request {
    return z.Request
}

// GetWriter returns http.ResponseWriter
func (z *baseContext) GetWriter() http.ResponseWriter {
    return z.Writer
}

func (z *baseContext) Write(data []byte) (int, error) {
    return z.Writer.Write(data)
}

func (z *baseContext) WriteWithCode(code int, data []byte) (int, error) {
    z.Writer.WriteHeader(code)
    return z.Write(data)
}
