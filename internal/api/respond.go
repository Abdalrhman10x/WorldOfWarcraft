package api

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"WorldOfWarcraft/internal/domain"
	"WorldOfWarcraft/internal/store"
)

type errorResponse struct {
	Error string `json:"error"`
}

func statusFor(err error) int {
	switch {
	case errors.Is(err, store.ErrNotFound):
		return http.StatusNotFound

	case errors.Is(err, store.ErrAlreadyExists),
		errors.Is(err, domain.ErrInventoryFull),
		errors.Is(err, domain.ErrNotEnoughGold),
		errors.Is(err, domain.ErrNotEnoughItems),
		errors.Is(err, domain.ErrIsAlreadyActive),
		errors.Is(err, domain.ErrIsAlreadySuspended),
		errors.Is(err, domain.ErrCharacterAlreadyReachedMaxLevel):
		return http.StatusConflict

	case errors.Is(err, domain.ErrInvalidCharacterName),
		errors.Is(err, domain.ErrUnknownRace),
		errors.Is(err, domain.ErrUnknownClass),
		errors.Is(err, domain.ErrInvalidClassCombo),
		errors.Is(err, domain.ErrUnknownItem),
		errors.Is(err, domain.ErrInvalidQuantity),
		errors.Is(err, domain.ErrInvalidXPAmount),
		errors.Is(err, domain.ErrInvalidAmount):
		return http.StatusBadRequest

	default:
		return http.StatusInternalServerError
	}
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		slog.Error("encode response", "error", err)
	}
}

func writeErr(w http.ResponseWriter, err error) {
	status := statusFor(err)

	if status >= 500 {
		slog.Error("request failed", "error", err)
		writeJSON(w, status, errorResponse{Error: "internal server error"})
		return
	}
	writeJSON(w, status, errorResponse{Error: err.Error()})
}
