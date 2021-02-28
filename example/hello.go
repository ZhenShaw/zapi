package main

import (
    "github.com/zhenshaw/go-lib/logs"
    "github.com/zhenshaw/zapi"
)

func main() {

    logs.CloseFileOutput()

    app := zapi.NewApp()
    r := app.GetRouter()

    r.Use(zapi.DefaultMiddlewares...)
    r.Add("/hello", &zapi.Context{}, Hello)

    r.Sub("/sub1").Add("/hello", &zapi.Context{}, Hello).Methods("GET")

    r.SubApi("/sub2", []*zapi.Api{
        r.NewApi("/hello", &zapi.Context{}, Hello).Methods("GET"),
        r.NewApi("/hello", &zapi.Context{}, Hello).Methods("POST"),
    })

    logs.Error(app.Run())
}

func Hello(z *zapi.Context) {
    z.Write([]byte("hello world"))
}
