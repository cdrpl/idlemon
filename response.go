package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
)

type ErrorResponse struct {
	Message string `json:"message"`
}

// Will write a successful json response. The data must be a valid target for JSON encoding.
func JsonRes(w http.ResponseWriter, data interface{}) {
	WriteJsonRes(w, http.StatusOK, data)
}

// Writes a json response {"status":0}
func JsonSuccess(w http.ResponseWriter) {
	WriteJsonRes(w, http.StatusOK, map[string]int{"status": 0})
}

// Writes an error response using the ResponseWriter.
func ErrRes(w http.ResponseWriter, code int) {
	e := ErrorResponse{Message: http.StatusText(code)}
	WriteJsonRes(w, code, e)
}

// Writes an error response using a custom message.
func ErrResCustom(w http.ResponseWriter, code int, msg string) {
	e := ErrorResponse{Message: msg}
	WriteJsonRes(w, code, e)
}

// Will write an error response with a custom message.
// If the ENV env var is set to production the message will be replaced with a standard one based on the HTTP code.
func ErrResSanitize(w http.ResponseWriter, code int, msg string) {
	e := ErrorResponse{}

	if os.Getenv("ENV") == "production" {
		e.Message = http.StatusText(code)
	} else {
		e.Message = msg
	}

	WriteJsonRes(w, code, e)
}

// Writes a json response.
func WriteJsonRes(w http.ResponseWriter, code int, data interface{}) {
	w.Header().Add("Content-Type", "application/json")

	w.WriteHeader(code)

	encoder := json.NewEncoder(w)
	err := encoder.Encode(data)
	if err != nil {
		log.Fatalf("failed to write json response: %v\n", err)
	}
}
