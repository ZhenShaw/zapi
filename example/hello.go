package main

import (
    "github.com/zhenshaw/go-lib/logs"
    "github.com/zhenshaw/zapi"
)

func main() {

    logs.CloseFileOutput()

    app := zapi.NewApp()
    r := app.GetRouter()

    r.Use(zapi.Recover, zapi.AccessLog)

    r.Add("/hello", &zapi.Context{}, Hello)

    r.Sub("/sub1").Add("/hello", &zapi.Context{}, Hello).Methods("GET")

    r.SubApi("/sub2", []*zapi.Api{
        r.NewApi("/hello", &zapi.Context{}, Hello).Methods("GET"),
        r.NewApi("/hello", &zapi.Context{}, Hello).Methods("POST"),
    })

    go func() {
        logs.Error(app.Run())
    }()

    app.Shutdown()
}

func Hello(z *zapi.Context) {
    z.Write([]byte("hello world"))
}
