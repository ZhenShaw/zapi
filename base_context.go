package zapi

import (
	"net/http"
	"reflect"
)

type IContext interface {
	Init(handlers []IHandler, w http.ResponseWriter, r *http.Request)
	Start(ctx IContext)
	Next()
	Finish()
	Reset()
	Switch(IHandler)
	GetRequest() *http.Request
	GetWriter() http.ResponseWriter
}

type BaseContext struct {
	Request *http.Request
	Writer  http.ResponseWriter

	handlers []IHandler

	// mark index in handlers chain
	index int

	ctx IContext
}

func (z *BaseContext) Init(handlers []IHandler, w http.ResponseWriter, r *http.Request) {
	z.handlers = handlers
	z.Writer = w
	z.Request = r
}

/*
Switch assert the exact argument type of a handler.
When implement BaseContext in a new struct, it must rewrite this function, for example:

func (c *MyCtx) Switch(handler IHandler) {
	switch handleFun := handler.(type) {
	case func(*MyCtx):
		handleFun(c)
	default:
		c.BaseContext.Switch(handler)
	}
}
*/
func (z *BaseContext) Switch(handler IHandler) {

	switch handleFun := handler.(type) {
	case func(*BaseContext):
		handleFun(z)
	case func(IContext):
		handleFun(z)
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
		ctx.Switch(z.handlers[0])
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
	z.ctx.Switch(handler)
}

// Finish allow you do some extra work after the last handler was executed.
func (z *BaseContext) Finish() {}

// GetWriter returns http.ResponseWriter
func (z *BaseContext) Reset() {
	z.Request = nil
	z.Writer = nil
	z.handlers = nil
}

// GetRequest returns *http.Request
func (z *BaseContext) GetRequest() *http.Request {
	return z.Request
}

// GetWriter returns http.ResponseWriter
func (z *BaseContext) GetWriter() http.ResponseWriter {
	return z.Writer
}

func (z *BaseContext) Write(data []byte) (int, error) {
	return z.Writer.Write(data)
}
