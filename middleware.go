package zapi

import (
    "fmt"
    "net/http"
    "runtime"

    "github.com/mo7zayed/reqip"
    "github.com/zhenshaw/go-lib/logs"
)

type MiddleWare func(*Context)

func Recover(c *Context) {
    defer func() {
        if err := recover(); err != nil {
            const size = 64 << 10
            buf := make([]byte, size)
            stack := buf[:runtime.Stack(buf, false)]
            err := fmt.Errorf("recover a panic: %v \n%s", err, stack)
            logs.Error(err.Error())
            c.Writer.WriteHeader(http.StatusInternalServerError)
            return
        }
    }()

    c.Next()
}

func AccessLog(c *Context) {

    ip := reqip.GetClientIP(c.Request)
    if ip == "" {
        ip = "unknown"
    }

    logs.Info("[ACCESS] %s => %s => %s", ip, c.Request.Method,
        c.Request.URL.Path)

    c.Next()
}

func Cors(c *Context) {
    r := c.Request
    w := c.Writer

    w.Header().Add("Access-Control-Allow-Origin", r.Header.Get("Origin"))
    w.Header().Add("Access-Control-Allow-Credentials", "true")
    w.Header().Add("Access-Control-Allow-Methods",
        "POST, GET, OPTIONS, PATCH, PUT, DELETE")

    if r.Method == http.MethodOptions {
        w.WriteHeader(http.StatusNoContent)
        return
    }

    c.Next()
}
