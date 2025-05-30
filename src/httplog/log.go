package httplog

import (
	"context"
	"fmt"
	"net/http"

	"github.com/noellimx/hepmilserver/src/controller/middlewares"

	"github.com/google/uuid"
)

func ContextualizeHttpRequest(req *http.Request) *http.Request {
	method := req.Method
	if req.Method == "" {
		method = "GET"
	}
	ctx := context.WithValue(req.Context(), "METHOD", method)

	traceId := uuid.New().String()
	ctx = context.WithValue(ctx, "TRACE_ID", traceId)
	ctx = context.WithValue(ctx, "URL/PATH", req.URL.String())
	ctx = context.WithValue(ctx, "ORIGIN", req.Header.Get("Origin"))
	ctx = context.WithValue(ctx, "USER-AGENT", req.Header.Get("User-Agent"))
	ctx = context.WithValue(ctx, "SESSION_ID", middlewares.GetSessionIdFromRequest(req))
	req = req.WithContext(ctx)
	return req
}

func SPrintHttpRequestPrefix(req *http.Request) string {
	ctx := req.Context()
	return fmt.Sprintf("[METHOD:%s URL:%s ORIGIN:%s UA: %s TRACE_ID:%s]", ctx.Value("METHOD"), ctx.Value("URL/PATH"), ctx.Value("ORIGIN"), ctx.Value("USER-AGENT"), ctx.Value("TRACE_ID"))
}
