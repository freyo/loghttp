package loghttp

import (
    "bytes"
    "context"
    "encoding/json"
    "fmt"
    "io/ioutil"
    "log"
    "net/http"
    "strings"
    "time"
)

// Transport implements http.RoundTripper. When set as Transport of http.Client, it executes HTTP requests with logging.
// No field is mandatory.
type Transport struct {
    Transport   http.RoundTripper
    LogRequest  func(req *http.Request) string
    LogResponse func(resp *http.Response) string
}

// THe default logging transport that wraps http.DefaultTransport.
var DefaultTransport = &Transport{
    Transport: http.DefaultTransport,
}

// Used if transport.LogRequest is not set.
var DefaultLogRequest = func(req *http.Request) string {
    header, _ := json.Marshal(req.Header)

    body, _ := ioutil.ReadAll(req.Body)
    req.Body = ioutil.NopCloser(bytes.NewBuffer(body))

    return fmt.Sprintf("--> %s %s %s %s", req.Method, req.URL, string(header), string(body))
}

// Used if transport.LogResponse is not set.
var DefaultLogResponse = func(resp *http.Response) string {
    ctx := resp.Request.Context()

    header, _ := json.Marshal(resp.Header)

    body, _ := ioutil.ReadAll(resp.Body)
    resp.Body = ioutil.NopCloser(bytes.NewBuffer(body))

    if start, ok := ctx.Value(ContextKeyRequestStart).(time.Time); ok {
        return fmt.Sprintf("<-- %d (%s) %s %s", resp.StatusCode, time.Since(start), string(header), strings.ReplaceAll(string(body), "\n", ""))
    } else {
        return fmt.Sprintf("<-- %d %s %s", resp.StatusCode, string(header), strings.ReplaceAll(string(body), "\n", ""))
    }
}

type contextKey struct {
    name string
}

var ContextKeyRequestStart = &contextKey{"RequestStart"}

// RoundTrip is the core part of this module and implements http.RoundTripper.
// Executes HTTP request with request/response logging.
func (t *Transport) RoundTrip(req *http.Request) (*http.Response, error) {
    ctx := context.WithValue(req.Context(), ContextKeyRequestStart, time.Now())
    req = req.WithContext(ctx)

    LogRequest := t.logRequest(req)

    resp, err := t.transport().RoundTrip(req)
    if err != nil {
        log.Printf("%s %s", LogRequest, err.Error())
        return resp, err
    }

    LogResponse := t.logResponse(resp)
    log.Printf("%s %s", LogRequest, LogResponse)

    return resp, err
}

func (t *Transport) logRequest(req *http.Request) string {
    if t.LogRequest != nil {
        return t.LogRequest(req)
    } else {
        return DefaultLogRequest(req)
    }
}

func (t *Transport) logResponse(resp *http.Response) string {
    if t.LogResponse != nil {
        return t.LogResponse(resp)
    } else {
        return DefaultLogResponse(resp)
    }
}

func (t *Transport) transport() http.RoundTripper {
    if t.Transport != nil {
        return t.Transport
    }

    return http.DefaultTransport
}
