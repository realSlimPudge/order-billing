package handler

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"order-billing/internal/order/model"
)

type errorResponse struct {
	Error string `json:"error"`
}

func respondJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		slog.Error("failed to encode response", "err", err)
	}
}

func respondError(w http.ResponseWriter, status int, message string) {
	respondJSON(w, status, errorResponse{Error: message})
}

func httpStatusOf(err error) int {
	switch {
	case errors.Is(err, model.ErrValidationFailed):
		return http.StatusBadRequest
	case errors.Is(err, model.ErrOrderNotFound):
		return http.StatusNotFound
	default:
		return http.StatusInternalServerError
	}
}
