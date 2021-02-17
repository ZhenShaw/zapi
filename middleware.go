package zapi

import (
	"fmt"
	"net/http"
	"runtime"

	"github.com/mo7zayed/reqip"
	"github.com/zhenshaw/go-lib/logs"
)

var DefaultMiddlewares = []IHandler{Recover, AccessLog, Cors}

func Recover(c IContext) {
	defer func() {
		if err := recover(); err != nil {
			const size = 64 << 10
			buf := make([]byte, size)
			stack := buf[:runtime.Stack(buf, false)]
			err := fmt.Errorf("recover a panic: %v \n%s", err, stack)
			logs.Error(err.Error())
			c.GetWriter().WriteHeader(http.StatusInternalServerError)
			return
		}
	}()
	c.Next()
}

func AccessLog(c IContext) {
	r := c.GetRequest()
	ip := reqip.GetClientIP(r)
	if ip == "" {
		ip = "unknown"
	}

	logs.Info("[ACCESS] %s => %s => %s", ip, r.Method, r.URL.Path)
	c.Next()
}

func Cors(c IContext) {
	r := c.GetRequest()
	w := c.GetWriter()

	w.Header().Add("Access-Control-Allow-Origin", r.Header.Get("Origin"))
	w.Header().Add("Access-Control-Allow-Credentials", "true")
	w.Header().Add("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PATCH, PUT, DELETE")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	c.Next()
}
