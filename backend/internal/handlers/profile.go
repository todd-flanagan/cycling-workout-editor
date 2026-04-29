package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/tflanaga/cycling-workout-editor/backend/internal/middleware"
	"github.com/tflanaga/cycling-workout-editor/backend/internal/models"
)

// GetProfile returns the current user's profile.
func (h *Handler) GetProfile(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		respondError(w, http.StatusUnauthorized, "not authenticated")
		return
	}

	user, err := h.UserRepo.GetByID(r.Context(), userID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "database error")
		return
	}
	if user == nil {
		respondError(w, http.StatusNotFound, "user not found")
		return
	}

	respondJSON(w, http.StatusOK, user)
}

// UpdateProfile updates the current user's profile.
func (h *Handler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		respondError(w, http.StatusUnauthorized, "not authenticated")
		return
	}

	var req models.ProfileUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	user, err := h.UserRepo.GetByID(r.Context(), userID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "database error")
		return
	}
	if user == nil {
		respondError(w, http.StatusNotFound, "user not found")
		return
	}

	if req.DisplayName != nil {
		if *req.DisplayName == "" {
			respondError(w, http.StatusBadRequest, "display_name cannot be empty")
			return
		}
		user.DisplayName = *req.DisplayName
	}
	if req.FTP != nil {
		if *req.FTP < 1 || *req.FTP > 1000 {
			respondError(w, http.StatusBadRequest, "ftp must be between 1 and 1000")
			return
		}
		user.FTP = *req.FTP
	}

	if err := h.UserRepo.Update(r.Context(), user); err != nil {
		respondError(w, http.StatusInternalServerError, "failed to update profile")
		return
	}

	respondJSON(w, http.StatusOK, user)
}
