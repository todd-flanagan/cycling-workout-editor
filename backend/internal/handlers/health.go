package handlers

import (
	"net/http"
)

// Healthz is the health check endpoint.
func (h *Handler) Healthz(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
