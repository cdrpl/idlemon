package main

import (
	"io"
	"log"
	"net/http"

	"github.com/google/uuid"
)

func ReadReqBody(w http.ResponseWriter, r *http.Request) ([]byte, error) {
	r.Body = http.MaxBytesReader(w, r.Body, MAX_REQ_BODY_SIZE)
	bytes, err := io.ReadAll(r.Body)
	if err != nil {
		return bytes, err
	}

	return bytes, nil
}

// Get the user ID from the request. Will call log.
func GetUserId(r *http.Request) uuid.UUID {
	id, ok := r.Context().Value(UserIdCtx).(uuid.UUID)
	if !ok {
		log.Println("failed to fetch user ID from request context") // this should never fail if auth middleware is correctly used
	}

	return id
}

// Get the request data object from the request.
func GetReqDto(r *http.Request) RequestDTO {
	dto, ok := r.Context().Value(ReqDtoCtx).(RequestDTO)
	if !ok {
		log.Println("failed to fetch DTO from request context") // this should never fail if body middleware is correctly used
	}

	return dto
}
