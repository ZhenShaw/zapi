package main

import (
	"github.com/zapi/zapi"
	"github.com/zhenshaw/go-lib/logs"
)

func main() {

	logs.CloseFileOutput()

	app := zapi.NewApp()
	r := app.GetRouter()

	r.Use(zapi.DefaultMiddlewares...)
	r.Add("/ping", &zapi.Context{}, Pong)

	logs.Error(app.Run())
}

func Pong(z *zapi.Context) {
	z.Write([]byte("hello world"))
}
