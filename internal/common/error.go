package common

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

type ErrorResponse struct {
	Error string `json:"error"`
}

var ErrorDecodeRequest = errors.New("decoding request is incorrect")

func MakeJSONError(w http.ResponseWriter, status int, err error) {
	w.WriteHeader(status)

	errorResponce := ErrorResponse{
		Error: err.Error(),
	}

	if err = json.NewEncoder(w).Encode(errorResponce); err != nil {
		fmt.Printf("error encoding response: %v\n", err)
	}
}
