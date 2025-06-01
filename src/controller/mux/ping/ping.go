package ping

import (
	"encoding/json"
	"github.com/noellimx/hepmilserver/src/controller/middlewares"
	"github.com/noellimx/hepmilserver/src/httplog"
	"log"
	"net/http"
)

type Response struct {
}

type PingHandler struct {
}

// PingHandler godoc
// @Summary      Ping the server.
// @Description  Returns a simple response to test connectivity
// @Tags         healthcheck
// @Accept       json
// @Produce      json
// @Success      200 {object} Response
// @Router       /ping [get]
func (h PingHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	sessionId := middlewares.GetSessionIdFromRequest(r)
	log.Printf("%s sessionId: %s\n", httplog.SPrintHttpRequestPrefix(r), sessionId)

	resp := Response{}
	respB, _ := json.Marshal(resp)
	log.Printf("%s response: %s\n", httplog.SPrintHttpRequestPrefix(r), string(respB))
	w.Write(respB)
}
