package middlewares

import (
	"net/http"
)

type Middleware = func(handler http.Handler) http.Handler
type MiddewareStack struct {
	h []Middleware
}

func (ms MiddewareStack) Wrap(next func(handler http.Handler) http.Handler) MiddewareStack {
	ms.h = append(ms.h, next)
	return ms
}

func (ms MiddewareStack) Finalize(final http.Handler) http.Handler {
	for i := range ms.h {
		final = ms.h[len(ms.h)-1-i](final)
	}
	return final
}
