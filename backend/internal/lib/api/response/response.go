package response

import (
	"encoding/json"
	"net/http"
)

type Response struct {
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
	Data    any    `json:"data,omitempty"`
}

const (
	StatusOK    = "ok"
	StatusError = "error"
)

func JSON(w http.ResponseWriter, code int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, `{"status":"error","message":"failed to encode response"}`, http.StatusInternalServerError)
	}
}

func OK(w http.ResponseWriter, data any) {
	JSON(w, http.StatusOK, Response{
		Status: StatusOK,
		Data:   data,
	})
}

func Created(w http.ResponseWriter, data any) {
	JSON(w, http.StatusCreated, Response{
		Status: StatusOK,
		Data:   data,
	})
}

func Error(w http.ResponseWriter, code int, message string) {
	JSON(w, code, Response{
		Status:  StatusError,
		Message: message,
	})
}

func BadRequest(w http.ResponseWriter, message string) {
	Error(w, http.StatusBadRequest, message)
}

func Unauthorized(w http.ResponseWriter) {
	Error(w, http.StatusUnauthorized, "unauthorized")
}

func Forbidden(w http.ResponseWriter) {
	Error(w, http.StatusForbidden, "forbidden")
}

func NotFound(w http.ResponseWriter) {
	Error(w, http.StatusNotFound, "not found")
}

func Internal(w http.ResponseWriter) {
	Error(w, http.StatusInternalServerError, "internal server error")
}
