package response_types

import (
	"encoding/json"
	"net/http"
)

type Response[T any] struct {
	Data  T       `json:"data"`
	Error *string `json:"error"`
}

func Error[T any](w http.ResponseWriter, httpCode int, err error, data T) {
	w.WriteHeader(httpCode)
	var r Response[T]

	r.Data = data
	if err != nil {
		msg := err.Error()
		r.Error = &msg
	}

	b, _ := json.Marshal(r)
	w.Write(b)
}

func ErrorNoBody(w http.ResponseWriter, httpCode int, err error) {
	Error[any](w, httpCode, err, nil)
}

func OkEmptyJsonBody(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("{}"))
}

func OkJsonBody[T any](w http.ResponseWriter, body T) {
	JsonBody(w, http.StatusOK, body)
}

func JsonBody[T any](w http.ResponseWriter, httpStatusCode int, body T) {
	w.Header().Set("Content-Type", "application/json")

	w.WriteHeader(httpStatusCode)
	var r Response[T]
	r.Data = body
	r.Error = nil
	b, _ := json.Marshal(r)
	w.Write(b)
}
