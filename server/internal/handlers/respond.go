// Package handlers holds JSON helpers shared by HTTP handlers.
package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"keeper/internal/logger"
)

const (
	MsgInvalidContentType = "Content-Type must be application/json"
	MsgInvalidBody        = "invalid request body"
	MsgInternalError      = "internal server error"
)

// ErrorResponse is the body of every error reply.
type ErrorResponse struct {
	Error string `json:"error"`
}

// DecodeJSON reads the request body into v or writes an error reply and returns false.
func DecodeJSON(w http.ResponseWriter, r *http.Request, v any) bool {
	if !strings.Contains(r.Header.Get("Content-Type"), "application/json") {
		RespondError(w, http.StatusUnsupportedMediaType, MsgInvalidContentType)
		return false
	}
	if err := json.NewDecoder(r.Body).Decode(v); err != nil {
		RespondError(w, http.StatusBadRequest, MsgInvalidBody)
		return false
	}
	return true
}

// RespondJSON writes payload as JSON with the given status code.
func RespondJSON(w http.ResponseWriter, code int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		logger.Error(err)
	}
}

// RespondError writes an error reply.
func RespondError(w http.ResponseWriter, code int, msg string) {
	RespondJSON(w, code, &ErrorResponse{Error: msg})
}

// RespondInternal logs err and writes an impersonal 500 reply.
func RespondInternal(w http.ResponseWriter, err error) {
	logger.Error(err)
	RespondError(w, http.StatusInternalServerError, MsgInternalError)
}
