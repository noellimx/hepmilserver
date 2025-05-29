package ping

import (
	"log"
	"net/http"

	"encoding/json"
	"github.com/noellimx/hepmilserver/src/controller/middlewares"
	"github.com/noellimx/hepmilserver/src/httplog"
)

type Response struct {
}

// PingHandler godoc
// @Summary      Ping the server
// @Description  Returns a simple response to test connectivity
// @Tags         health
// @Accept       json
// @Produce      json
// @Success      200 {object} Response
// @Router       /ping [get]
func PingHandler(w http.ResponseWriter, r *http.Request) {
	sessionId := middlewares.GetSessionIdFromRequest(r)
	log.Printf("%s sessionId: %s\n", httplog.SPrintHttpRequestPrefix(r), sessionId)

	resp := Response{}

	respB, _ := json.Marshal(resp)

	log.Printf("%s response: %s\n", httplog.SPrintHttpRequestPrefix(r), string(respB))
	w.Write(respB)
}

func Register(mux *http.ServeMux, mws middlewares.MiddewareStack) {
	// Checks current user state of the client
	// Provides server configuration values

	mux.Handle("/ping", mws.Finalize(http.HandlerFunc(PingHandler)))

}
