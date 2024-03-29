package zapi

import (
    "fmt"
    "net/http"
    "runtime"
    "time"

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
    c.Next()

    ip := reqip.GetClientIP(c.Request)
    elapsed := time.Since(c.Begin)
    logs.Info("[ACCESS] %d => %s => %s %s %s", c.Writer.Status, elapsed,
        ip, c.Request.Method, c.Request.URL.String())

}
