package utils

import (
	"encoding/json"
	"net/http"
)

type JsonResponse struct {
	Message    string      `json:"message"`
	StatusCode int         `json:"statusCode"`
	Error      interface{} `json:"error"`
	Data       interface{} `json:"data"`
}

func WriteJson(w http.ResponseWriter, status int, message string, data interface{}, err interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	response := JsonResponse{
		Message:    message,
		StatusCode: status,
		Error:      err,
		Data:       data,
	}
	json.NewEncoder(w).Encode(response)
}
