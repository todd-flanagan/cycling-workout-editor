package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/tflanaga/cycling-workout-editor/backend/internal/auth"
	"github.com/tflanaga/cycling-workout-editor/backend/internal/repository"
)

// Handler groups all HTTP handler dependencies.
type Handler struct {
	UserRepo        repository.UserRepository
	FTPHistoryRepo  repository.FTPHistoryRepository
	WorkoutRepo     repository.WorkoutRepository
	JWTManager      *auth.JWTManager
	OAuthConfig     OAuthProvider
	FrontendURL     string
	userInfoFetcher UserInfoFetcher
}

// NewHandler creates a new Handler with all dependencies.
func NewHandler(
	userRepo repository.UserRepository,
	ftpHistoryRepo repository.FTPHistoryRepository,
	workoutRepo repository.WorkoutRepository,
	jwtManager *auth.JWTManager,
	oauthProvider OAuthProvider,
	frontendURL string,
) *Handler {
	return &Handler{
		UserRepo:       userRepo,
		FTPHistoryRepo: ftpHistoryRepo,
		WorkoutRepo:    workoutRepo,
		JWTManager:     jwtManager,
		OAuthConfig:    oauthProvider,
		FrontendURL:    frontendURL,
	}
}

// respondJSON writes a JSON response with the given status code.
func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if data != nil {
		json.NewEncoder(w).Encode(data)
	}
}

// respondError writes a JSON error response.
func respondError(w http.ResponseWriter, status int, msg string) {
	respondJSON(w, status, map[string]string{"error": msg})
}
