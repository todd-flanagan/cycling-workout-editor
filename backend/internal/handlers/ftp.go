package handlers

import (
	"encoding/json"
	"math"
	"net/http"

	"github.com/google/uuid"
	"github.com/tflanaga/cycling-workout-editor/backend/internal/middleware"
	"github.com/tflanaga/cycling-workout-editor/backend/internal/models"
)

// GetFTPHistory returns the FTP history for the current user.
func (h *Handler) GetFTPHistory(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		respondError(w, http.StatusUnauthorized, "not authenticated")
		return
	}

	entries, err := h.FTPHistoryRepo.List(r.Context(), userID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "database error")
		return
	}

	if entries == nil {
		entries = []models.FTPHistory{}
	}

	respondJSON(w, http.StatusOK, entries)
}

// AddFTPHistory adds a new FTP history entry.
func (h *Handler) AddFTPHistory(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		respondError(w, http.StatusUnauthorized, "not authenticated")
		return
	}

	var req models.AddFTPHistoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.FTP < 1 || req.FTP > 1000 {
		respondError(w, http.StatusBadRequest, "ftp must be between 1 and 1000")
		return
	}
	if req.RecordedAt == "" {
		respondError(w, http.StatusBadRequest, "recorded_at is required")
		return
	}

	entry := &models.FTPHistory{
		ID:         uuid.New().String(),
		UserID:     userID,
		FTP:        req.FTP,
		RecordedAt: req.RecordedAt,
		Source:     req.Source,
	}

	if err := h.FTPHistoryRepo.Create(r.Context(), entry); err != nil {
		respondError(w, http.StatusInternalServerError, "failed to add FTP history entry")
		return
	}

	respondJSON(w, http.StatusCreated, entry)
}

// FTPFromRamp calculates FTP from max 1-minute ramp power (75%).
// If confirm is true, it also saves the FTP to history and updates the user's current FTP.
func (h *Handler) FTPFromRamp(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		respondError(w, http.StatusUnauthorized, "not authenticated")
		return
	}

	var req models.FTPFromRampRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.MaxOnMinPower < 1 {
		respondError(w, http.StatusBadRequest, "max_one_min_power must be positive")
		return
	}

	estimatedFTP := int(math.Round(float64(req.MaxOnMinPower) * 0.75))

	resp := models.FTPFromRampResponse{
		EstimatedFTP: estimatedFTP,
		Confirmed:    false,
	}

	if req.Confirm {
		// Save to FTP history.
		entry := &models.FTPHistory{
			ID:         uuid.New().String(),
			UserID:     userID,
			FTP:        estimatedFTP,
			RecordedAt: r.URL.Query().Get("date"), // Optional date override.
			Source:     "ramp test",
		}
		if entry.RecordedAt == "" {
			entry.RecordedAt = "2025-01-01" // Fallback; in practice the frontend sends a date.
		}

		if err := h.FTPHistoryRepo.Create(r.Context(), entry); err != nil {
			respondError(w, http.StatusInternalServerError, "failed to save FTP history")
			return
		}

		// Update user's current FTP.
		user, err := h.UserRepo.GetByID(r.Context(), userID)
		if err != nil || user == nil {
			respondError(w, http.StatusInternalServerError, "failed to get user")
			return
		}
		user.FTP = estimatedFTP
		if err := h.UserRepo.Update(r.Context(), user); err != nil {
			respondError(w, http.StatusInternalServerError, "failed to update user FTP")
			return
		}

		resp.Confirmed = true
	}

	respondJSON(w, http.StatusOK, resp)
}
