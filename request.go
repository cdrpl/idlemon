package main

import (
	"io"
	"net/http"
)

func ReadReqBody(w http.ResponseWriter, r *http.Request) ([]byte, error) {
	r.Body = http.MaxBytesReader(w, r.Body, MAX_REQ_BODY_SIZE)
	bytes, err := io.ReadAll(r.Body)
	if err != nil {
		return bytes, err
	}

	return bytes, nil
}

// Get the user ID from the request.
func GetUserID(r *http.Request) int {
	return r.Context().Value(UserIdCtx).(int)
}

// Get the request data object from the request.
func GetReqDTO(r *http.Request) RequestDTO {
	return r.Context().Value(ReqDtoCtx).(RequestDTO)
}
