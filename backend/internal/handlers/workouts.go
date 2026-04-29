package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/tflanaga/cycling-workout-editor/backend/internal/middleware"
	"github.com/tflanaga/cycling-workout-editor/backend/internal/models"
)

// CreateWorkout creates a new workout for the current user.
func (h *Handler) CreateWorkout(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		respondError(w, http.StatusUnauthorized, "not authenticated")
		return
	}

	var req models.CreateWorkoutRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Name == "" {
		req.Name = "Untitled Workout"
	}
	if req.Intervals == nil {
		req.Intervals = []models.Interval{}
	}

	workout := &models.Workout{
		ID:          uuid.New().String(),
		UserID:      userID,
		Name:        req.Name,
		Description: req.Description,
		Intervals:   req.Intervals,
	}

	if err := h.WorkoutRepo.Create(r.Context(), workout); err != nil {
		respondError(w, http.StatusInternalServerError, "failed to create workout")
		return
	}

	respondJSON(w, http.StatusCreated, workout)
}

// GetWorkout returns a single workout by ID, scoped to the current user.
func (h *Handler) GetWorkout(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		respondError(w, http.StatusUnauthorized, "not authenticated")
		return
	}

	workoutID := chi.URLParam(r, "id")
	if workoutID == "" {
		respondError(w, http.StatusBadRequest, "missing workout id")
		return
	}

	workout, err := h.WorkoutRepo.GetByID(r.Context(), workoutID, userID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "database error")
		return
	}
	if workout == nil {
		respondError(w, http.StatusNotFound, "workout not found")
		return
	}

	respondJSON(w, http.StatusOK, workout)
}

// UpdateWorkout updates an existing workout by ID, scoped to the current user.
func (h *Handler) UpdateWorkout(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		respondError(w, http.StatusUnauthorized, "not authenticated")
		return
	}

	workoutID := chi.URLParam(r, "id")
	if workoutID == "" {
		respondError(w, http.StatusBadRequest, "missing workout id")
		return
	}

	workout, err := h.WorkoutRepo.GetByID(r.Context(), workoutID, userID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "database error")
		return
	}
	if workout == nil {
		respondError(w, http.StatusNotFound, "workout not found")
		return
	}

	var req models.UpdateWorkoutRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Name != nil {
		workout.Name = *req.Name
	}
	if req.Description != nil {
		workout.Description = *req.Description
	}
	if req.Intervals != nil {
		workout.Intervals = req.Intervals
	}

	if err := h.WorkoutRepo.Update(r.Context(), workout); err != nil {
		respondError(w, http.StatusInternalServerError, "failed to update workout")
		return
	}

	respondJSON(w, http.StatusOK, workout)
}
