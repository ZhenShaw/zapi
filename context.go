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
	Implement and rewrite.

	Init()
	Handel()
	Reset()
*/

func (c *Context) Init(handlers []IHandler, w http.ResponseWriter, r *http.Request) {
	c.BaseContext.Init(handlers, w, r)
}

func (c *Context) Handle(handler IHandler) {
	switch handleFun := handler.(type) {
	case func(*Context):
		handleFun(c)
	default:
		c.BaseContext.Handle(handler)
	}
}

func (c *Context) Reset() {
	c.BaseContext.Reset()
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

func (c *Context) CheckBind(obj interface{}) error {

	if err := c.Bind(obj); err != nil {
		return err
	}

	if _, err := govalidator.ValidateStruct(obj); err != nil {
		return err
	}
	return nil
}
