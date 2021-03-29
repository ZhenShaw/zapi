package zapi

import (
    "net/http"
    "time"

    "github.com/gin-gonic/gin/binding"
)

type Context struct {
    baseContext
    Begin time.Time
}

/*
	Implement and rewrite functions.
*/

func (c *Context) Init(handlers []IHandler, w http.ResponseWriter, r *http.Request) {
    c.baseContext.Init(handlers, w, r)
    // init other fields below if has.
    c.Begin = time.Now()
}

func (c *Context) Reset() {
    c.baseContext.Reset()
    // reset other fields below if has.
}

func (c *Context) Call(handler IHandler) {
    switch handleFun := handler.(type) {
    case func(*Context):
        handleFun(c)
    case MiddleWare:
        handleFun(c)
    default:
        c.baseContext.Call(handler)
    }
}

/*
	Extend functions.
*/

func (c *Context) ContentType() string {
    content := c.Request.Header.Get("Content-Type")
    for i, char := range content {
        if char == ' ' || char == ';' {
            return content[:i]
        }
    }
    return content
}

func (c *Context) Bind(obj interface{}) error {
    bind := binding.Default(c.Request.Method, c.ContentType())
    return bind.Bind(c.Request, obj)
}

func (c *Context) BindQuery(obj interface{}) error {
    bind := binding.Query
    return bind.Bind(c.Request, obj)
}

func (c *Context) RequestWithVars(vars map[string]string) {
    c.Request = requestWithVars(c.Request, vars)
}
