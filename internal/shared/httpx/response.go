package httpx

import (
	"encoding/json"
	"net/http"
)

type Meta struct {
	Page  int `json:"page"`
	Limit int `json:"limit"`
	Total int `json:"total"`
}

type successEnvelope struct {
	Data any   `json:"data"`
	Meta *Meta `json:"meta,omitempty"`
}

type ErrorDetail struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

type errorBody struct {
	Code    string        `json:"code"`
	Message string        `json:"message"`
	Details []ErrorDetail `json:"details,omitempty"`
}

type errorEnvelope struct {
	Error errorBody `json:"error"`
}

func JSON(w http.ResponseWriter, status int, data any) {
	write(w, status, successEnvelope{Data: data})
}

func JSONWithMeta(w http.ResponseWriter, status int, data any, meta Meta) {
	write(w, status, successEnvelope{Data: data, Meta: &meta})
}

func Error(w http.ResponseWriter, status int, code, message string) {
	write(w, status, errorEnvelope{Error: errorBody{Code: code, Message: message}})
}

func ValidationError(w http.ResponseWriter, message string, details []ErrorDetail) {
	write(w, http.StatusBadRequest, errorEnvelope{
		Error: errorBody{
			Code:    "VALIDACAO_INVALIDA",
			Message: message,
			Details: details,
		},
	})
}

func write(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}
