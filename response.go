package zapi

import (
    "bufio"
    "net"
    "net/http"
)

type Response struct {
    http.ResponseWriter
    http.Hijacker
    Responded bool
    Status    int
}

func (r *Response) reset(w http.ResponseWriter) {
    r.ResponseWriter = w
    r.Status = 0
    r.Responded = false
}

// Write writes the data to the connection as part of a HTTP reply,
// and sets `Responded` to true.
// Responded:  if true, the response was already sent
func (r *Response) Write(p []byte) (int, error) {
    if r.Status == 0 {
        r.Status = http.StatusOK
    }
    r.Responded = true
    return r.ResponseWriter.Write(p)
}

// WriteHeader sends a HTTP response header with status code,
// and sets `Responded` to true.
func (r *Response) WriteHeader(code int) {
    if r.Status > 0 {
        // prevent multiple response.WriteHeader calls
        return
    }
    r.Status = code
    r.Responded = true
    r.ResponseWriter.WriteHeader(code)
}

// Hijack implements the http.Hijacker interface.
func (r *Response) Hijack() (net.Conn, *bufio.ReadWriter, error) {
    return r.ResponseWriter.(http.Hijacker).Hijack()
}