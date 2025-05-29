package ping

import (
	"log"
	"net/http"

	"encoding/json"
	"github.com/noellimx/hepmilserver/src/controller/middlewares"
	"github.com/noellimx/hepmilserver/src/httplog"
)

type Response struct {
	LoginUrls map[string]string `json:"login_urls"`
}

func Register(mux *http.ServeMux, goauthloginurl string, mws middlewares.MiddewareStack) {
	// Checks current user state of the client
	// Provides server configuration values

	mux.Handle("/ping", mws.Finalize(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sessionId := middlewares.GetSessionIdFromRequest(r)
		log.Printf("%s sessionId: %s\n", httplog.SPrintHttpRequestPrefix(r), sessionId)

		resp := Response{LoginUrls: make(map[string]string)}
		resp.LoginUrls["google"] = goauthloginurl

		respB, _ := json.Marshal(resp)

		log.Printf("%s response: %s\n", httplog.SPrintHttpRequestPrefix(r), string(respB))
		w.Write(respB)
	})))

}
