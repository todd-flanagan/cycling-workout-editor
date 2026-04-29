package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/tflanaga/cycling-workout-editor/backend/internal/auth"
	"github.com/tflanaga/cycling-workout-editor/backend/internal/db"
	"github.com/tflanaga/cycling-workout-editor/backend/internal/handlers"
	"github.com/tflanaga/cycling-workout-editor/backend/internal/middleware"
	"github.com/tflanaga/cycling-workout-editor/backend/internal/repository"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

func main() {
	// --- Configuration from environment ---
	port := envOrDefault("PORT", "8080")
	dbPath := envOrDefault("DATABASE_PATH", "./data/workouts.db")
	jwtSecret := envOrDefault("JWT_SECRET", "change-me-in-production")
	jwtExpiryStr := envOrDefault("JWT_EXPIRY", "24h")
	googleClientID := envOrDefault("GOOGLE_CLIENT_ID", "")
	googleClientSecret := envOrDefault("GOOGLE_CLIENT_SECRET", "")
	frontendURL := envOrDefault("FRONTEND_URL", "http://localhost:3000")

	jwtExpiry, err := time.ParseDuration(jwtExpiryStr)
	if err != nil {
		log.Fatalf("invalid JWT_EXPIRY: %v", err)
	}

	// --- Database ---
	database, err := db.Open(dbPath)
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
	}
	defer database.Close()

	if err := db.RunMigrations(database); err != nil {
		log.Fatalf("failed to run migrations: %v", err)
	}

	// --- Repositories ---
	userRepo := repository.NewSQLiteUserRepo(database)
	ftpHistoryRepo := repository.NewSQLiteFTPHistoryRepo(database)
	workoutRepo := repository.NewSQLiteWorkoutRepo(database)

	// --- Auth ---
	jwtManager := auth.NewJWTManager(jwtSecret, jwtExpiry)

	baseURL := envOrDefault("BASE_URL", "http://localhost:8080")
	oauthConfig := &oauth2.Config{
		ClientID:     googleClientID,
		ClientSecret: googleClientSecret,
		RedirectURL:  fmt.Sprintf("%s/auth/google/callback", baseURL),
		Scopes:       []string{"openid", "email", "profile"},
		Endpoint:     google.Endpoint,
	}

	// Wrap oauth2.Config to satisfy the OAuthProvider interface.
	oauthProvider := &oauthConfigWrapper{Config: oauthConfig}

	// --- Handlers ---
	h := handlers.NewHandler(userRepo, ftpHistoryRepo, workoutRepo, jwtManager, oauthProvider, frontendURL)

	// --- Router ---
	r := chi.NewRouter()

	// Global middleware.
	r.Use(chimw.Logger)
	r.Use(chimw.Recoverer)
	r.Use(chimw.RequestID)
	r.Use(corsMiddleware(frontendURL))

	// Public routes.
	r.Get("/healthz", h.Healthz)
	r.Get("/auth/google/login", h.GoogleLogin)
	r.Get("/auth/google/callback", h.GoogleCallback)

	// Protected API routes.
	r.Route("/api", func(r chi.Router) {
		r.Use(middleware.JWTAuth(jwtManager))

		// User profile.
		r.Get("/user/profile", h.GetProfile)
		r.Put("/user/profile", h.UpdateProfile)

		// FTP.
		r.Get("/user/ftp-history", h.GetFTPHistory)
		r.Post("/user/ftp-history", h.AddFTPHistory)
		r.Post("/user/ftp-from-ramp", h.FTPFromRamp)

		// Workouts.
		r.Post("/workouts", h.CreateWorkout)
		r.Get("/workouts/{id}", h.GetWorkout)
		r.Put("/workouts/{id}", h.UpdateWorkout)
	})

	// --- Start server ---
	addr := ":" + port
	log.Printf("Server starting on %s", addr)
	if err := http.ListenAndServe(addr, r); err != nil {
		log.Fatalf("server error: %v", err)
	}
}

func envOrDefault(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}

// corsMiddleware returns a simple CORS middleware for the given origin.
func corsMiddleware(allowedOrigin string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", allowedOrigin)
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
			w.Header().Set("Access-Control-Allow-Credentials", "true")

			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusOK)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// oauthConfigWrapper wraps *oauth2.Config to implement the handlers.OAuthProvider interface.
type oauthConfigWrapper struct {
	Config *oauth2.Config
}

func (w *oauthConfigWrapper) AuthCodeURL(state string, opts ...oauth2.AuthCodeOption) string {
	return w.Config.AuthCodeURL(state, opts...)
}

func (w *oauthConfigWrapper) Exchange(ctx context.Context, code string, opts ...oauth2.AuthCodeOption) (*oauth2.Token, error) {
	return w.Config.Exchange(ctx, code, opts...)
}

func (w *oauthConfigWrapper) Client(ctx context.Context, token *oauth2.Token) *http.Client {
	return w.Config.Client(ctx, token)
}
