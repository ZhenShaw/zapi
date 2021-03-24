package zapi

import (
    "net/http"

    "github.com/asaskevich/govalidator"
    "github.com/gin-gonic/gin/binding"
)

type Context struct {
    BaseContext
}

/*
	Implement and rewrite functions.
*/

func (c *Context) Init(handlers []IHandler, w http.ResponseWriter, r *http.Request) {
    c.BaseContext.Init(handlers, w, r)
    // init other fields below if has.
}

func (c *Context) Reset() {
    c.BaseContext.Reset()
    // reset other fields below if has.
}

func (c *Context) Call(handler IHandler) {
    switch handleFun := handler.(type) {
    case func(*Context):
        handleFun(c)
    default:
        c.BaseContext.Call(handler)
    }
}

/*
	Extend functions.
*/

func (z *BaseContext) Write(data []byte) (int, error) {
    return z.Writer.Write(data)
}

func (z *BaseContext) WriteWithCode(code int, data []byte) (int, error) {
    z.Writer.WriteHeader(code)
    return z.Write(data)
}

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

func (c *Context) CheckBind(obj interface{}) error {

    if err := c.Bind(obj); err != nil {
        return err
    }

    if _, err := govalidator.ValidateStruct(obj); err != nil {
        return err
    }
    return nil
}
