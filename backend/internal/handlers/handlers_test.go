package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/tflanaga/cycling-workout-editor/backend/internal/auth"
	"github.com/tflanaga/cycling-workout-editor/backend/internal/middleware"
	"github.com/tflanaga/cycling-workout-editor/backend/internal/models"
)

func newTestHandler() (*Handler, *mockUserRepo, *mockFTPHistoryRepo, *mockWorkoutRepo) {
	userRepo := newMockUserRepo()
	ftpRepo := newMockFTPHistoryRepo()
	workoutRepo := newMockWorkoutRepo()
	jwtMgr := auth.NewJWTManager("test-handler-secret", 24*time.Hour)
	oauthProvider := newMockOAuthProvider()

	h := NewHandler(userRepo, ftpRepo, workoutRepo, jwtMgr, oauthProvider, "http://localhost:3000")
	return h, userRepo, ftpRepo, workoutRepo
}

// requestWithUser creates a request with the user ID injected into context
// (simulating what the auth middleware does).
func requestWithUser(method, path string, body []byte, userID string) *http.Request {
	var req *http.Request
	if body != nil {
		req = httptest.NewRequest(method, path, bytes.NewReader(body))
	} else {
		req = httptest.NewRequest(method, path, nil)
	}
	req.Header.Set("Content-Type", "application/json")
	ctx := context.WithValue(req.Context(), middleware.UserIDKey, userID)
	return req.WithContext(ctx)
}

// --- Health ---

func TestHealthz(t *testing.T) {
	h, _, _, _ := newTestHandler()
	req := httptest.NewRequest("GET", "/healthz", nil)
	rr := httptest.NewRecorder()

	h.Healthz(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rr.Code)
	}

	var body map[string]string
	json.NewDecoder(rr.Body).Decode(&body)
	if body["status"] != "ok" {
		t.Errorf("status: got %q, want %q", body["status"], "ok")
	}
}

// --- Profile ---

func TestGetProfile(t *testing.T) {
	h, userRepo, _, _ := newTestHandler()

	// Seed user.
	userRepo.Create(context.Background(), &models.User{
		ID: "user-1", GoogleID: "g1", Email: "a@b.com", DisplayName: "Alice", FTP: 250,
	})

	req := requestWithUser("GET", "/api/user/profile", nil, "user-1")
	rr := httptest.NewRecorder()

	h.GetProfile(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rr.Code)
	}

	var user models.User
	json.NewDecoder(rr.Body).Decode(&user)
	if user.Email != "a@b.com" {
		t.Errorf("email: got %q, want %q", user.Email, "a@b.com")
	}
	if user.FTP != 250 {
		t.Errorf("ftp: got %d, want 250", user.FTP)
	}
}

func TestGetProfileNotAuthenticated(t *testing.T) {
	h, _, _, _ := newTestHandler()

	req := httptest.NewRequest("GET", "/api/user/profile", nil)
	rr := httptest.NewRecorder()

	h.GetProfile(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", rr.Code)
	}
}

func TestGetProfileNotFound(t *testing.T) {
	h, _, _, _ := newTestHandler()

	req := requestWithUser("GET", "/api/user/profile", nil, "nonexistent")
	rr := httptest.NewRecorder()

	h.GetProfile(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", rr.Code)
	}
}

func TestUpdateProfile(t *testing.T) {
	h, userRepo, _, _ := newTestHandler()

	userRepo.Create(context.Background(), &models.User{
		ID: "user-1", GoogleID: "g1", Email: "a@b.com", DisplayName: "Alice", FTP: 200,
	})

	body, _ := json.Marshal(models.ProfileUpdateRequest{
		FTP: intPtr(275),
	})
	req := requestWithUser("PUT", "/api/user/profile", body, "user-1")
	rr := httptest.NewRecorder()

	h.UpdateProfile(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rr.Code)
	}

	var user models.User
	json.NewDecoder(rr.Body).Decode(&user)
	if user.FTP != 275 {
		t.Errorf("ftp: got %d, want 275", user.FTP)
	}
}

func TestUpdateProfileDisplayName(t *testing.T) {
	h, userRepo, _, _ := newTestHandler()

	userRepo.Create(context.Background(), &models.User{
		ID: "user-1", GoogleID: "g1", Email: "a@b.com", DisplayName: "Alice", FTP: 200,
	})

	name := "Bob"
	body, _ := json.Marshal(models.ProfileUpdateRequest{
		DisplayName: &name,
	})
	req := requestWithUser("PUT", "/api/user/profile", body, "user-1")
	rr := httptest.NewRecorder()

	h.UpdateProfile(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rr.Code)
	}

	var user models.User
	json.NewDecoder(rr.Body).Decode(&user)
	if user.DisplayName != "Bob" {
		t.Errorf("display_name: got %q, want %q", user.DisplayName, "Bob")
	}
}

func TestUpdateProfileInvalidFTP(t *testing.T) {
	h, userRepo, _, _ := newTestHandler()

	userRepo.Create(context.Background(), &models.User{
		ID: "user-1", GoogleID: "g1", Email: "a@b.com", DisplayName: "Alice", FTP: 200,
	})

	body, _ := json.Marshal(models.ProfileUpdateRequest{
		FTP: intPtr(0),
	})
	req := requestWithUser("PUT", "/api/user/profile", body, "user-1")
	rr := httptest.NewRecorder()

	h.UpdateProfile(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rr.Code)
	}
}

func TestUpdateProfileEmptyDisplayName(t *testing.T) {
	h, userRepo, _, _ := newTestHandler()

	userRepo.Create(context.Background(), &models.User{
		ID: "user-1", GoogleID: "g1", Email: "a@b.com", DisplayName: "Alice", FTP: 200,
	})

	empty := ""
	body, _ := json.Marshal(models.ProfileUpdateRequest{
		DisplayName: &empty,
	})
	req := requestWithUser("PUT", "/api/user/profile", body, "user-1")
	rr := httptest.NewRecorder()

	h.UpdateProfile(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rr.Code)
	}
}

func TestUpdateProfileInvalidBody(t *testing.T) {
	h, userRepo, _, _ := newTestHandler()

	userRepo.Create(context.Background(), &models.User{
		ID: "user-1", GoogleID: "g1", Email: "a@b.com", DisplayName: "Alice", FTP: 200,
	})

	req := requestWithUser("PUT", "/api/user/profile", []byte("not json"), "user-1")
	rr := httptest.NewRecorder()

	h.UpdateProfile(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rr.Code)
	}
}

// --- FTP ---

func TestGetFTPHistory(t *testing.T) {
	h, userRepo, ftpRepo, _ := newTestHandler()

	userRepo.Create(context.Background(), &models.User{
		ID: "user-1", GoogleID: "g1", Email: "a@b.com", DisplayName: "A", FTP: 200,
	})

	ftpRepo.Create(context.Background(), &models.FTPHistory{
		ID: "ftp-1", UserID: "user-1", FTP: 250, RecordedAt: "2025-06-15", Source: "ramp test",
	})

	req := requestWithUser("GET", "/api/user/ftp-history", nil, "user-1")
	rr := httptest.NewRecorder()

	h.GetFTPHistory(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rr.Code)
	}

	var entries []models.FTPHistory
	json.NewDecoder(rr.Body).Decode(&entries)
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	if entries[0].FTP != 250 {
		t.Errorf("ftp: got %d, want 250", entries[0].FTP)
	}
}

func TestGetFTPHistoryEmpty(t *testing.T) {
	h, userRepo, _, _ := newTestHandler()

	userRepo.Create(context.Background(), &models.User{
		ID: "user-1", GoogleID: "g1", Email: "a@b.com", DisplayName: "A", FTP: 200,
	})

	req := requestWithUser("GET", "/api/user/ftp-history", nil, "user-1")
	rr := httptest.NewRecorder()

	h.GetFTPHistory(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rr.Code)
	}

	var entries []models.FTPHistory
	json.NewDecoder(rr.Body).Decode(&entries)
	if len(entries) != 0 {
		t.Errorf("expected empty array, got %d entries", len(entries))
	}
}

func TestAddFTPHistory(t *testing.T) {
	h, userRepo, _, _ := newTestHandler()

	userRepo.Create(context.Background(), &models.User{
		ID: "user-1", GoogleID: "g1", Email: "a@b.com", DisplayName: "A", FTP: 200,
	})

	body, _ := json.Marshal(models.AddFTPHistoryRequest{
		FTP: 260, RecordedAt: "2025-06-15", Source: "20-min test",
	})
	req := requestWithUser("POST", "/api/user/ftp-history", body, "user-1")
	rr := httptest.NewRecorder()

	h.AddFTPHistory(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d", rr.Code)
	}

	var entry models.FTPHistory
	json.NewDecoder(rr.Body).Decode(&entry)
	if entry.FTP != 260 {
		t.Errorf("ftp: got %d, want 260", entry.FTP)
	}
	if entry.ID == "" {
		t.Error("expected non-empty ID")
	}
}

func TestAddFTPHistoryInvalidFTP(t *testing.T) {
	h, _, _, _ := newTestHandler()

	body, _ := json.Marshal(models.AddFTPHistoryRequest{
		FTP: 0, RecordedAt: "2025-06-15",
	})
	req := requestWithUser("POST", "/api/user/ftp-history", body, "user-1")
	rr := httptest.NewRecorder()

	h.AddFTPHistory(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rr.Code)
	}
}

func TestAddFTPHistoryMissingDate(t *testing.T) {
	h, _, _, _ := newTestHandler()

	body, _ := json.Marshal(models.AddFTPHistoryRequest{
		FTP: 250,
	})
	req := requestWithUser("POST", "/api/user/ftp-history", body, "user-1")
	rr := httptest.NewRecorder()

	h.AddFTPHistory(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rr.Code)
	}
}

func TestFTPFromRampCalculation(t *testing.T) {
	h, userRepo, _, _ := newTestHandler()

	userRepo.Create(context.Background(), &models.User{
		ID: "user-1", GoogleID: "g1", Email: "a@b.com", DisplayName: "A", FTP: 200,
	})

	body, _ := json.Marshal(models.FTPFromRampRequest{
		MaxOnMinPower: 400,
		Confirm:       false,
	})
	req := requestWithUser("POST", "/api/user/ftp-from-ramp", body, "user-1")
	rr := httptest.NewRecorder()

	h.FTPFromRamp(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rr.Code)
	}

	var resp models.FTPFromRampResponse
	json.NewDecoder(rr.Body).Decode(&resp)
	if resp.EstimatedFTP != 300 {
		t.Errorf("estimated ftp: got %d, want 300", resp.EstimatedFTP)
	}
	if resp.Confirmed {
		t.Error("expected confirmed=false")
	}
}

func TestFTPFromRampConfirm(t *testing.T) {
	h, userRepo, _, _ := newTestHandler()

	userRepo.Create(context.Background(), &models.User{
		ID: "user-1", GoogleID: "g1", Email: "a@b.com", DisplayName: "A", FTP: 200,
	})

	body, _ := json.Marshal(models.FTPFromRampRequest{
		MaxOnMinPower: 400,
		Confirm:       true,
	})
	req := requestWithUser("POST", "/api/user/ftp-from-ramp", body, "user-1")
	rr := httptest.NewRecorder()

	h.FTPFromRamp(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rr.Code)
	}

	var resp models.FTPFromRampResponse
	json.NewDecoder(rr.Body).Decode(&resp)
	if resp.EstimatedFTP != 300 {
		t.Errorf("estimated ftp: got %d, want 300", resp.EstimatedFTP)
	}
	if !resp.Confirmed {
		t.Error("expected confirmed=true")
	}

	// Verify user FTP was updated.
	u, _ := userRepo.GetByID(context.Background(), "user-1")
	if u.FTP != 300 {
		t.Errorf("user FTP: got %d, want 300", u.FTP)
	}
}

func TestFTPFromRampInvalidPower(t *testing.T) {
	h, _, _, _ := newTestHandler()

	body, _ := json.Marshal(models.FTPFromRampRequest{
		MaxOnMinPower: 0,
	})
	req := requestWithUser("POST", "/api/user/ftp-from-ramp", body, "user-1")
	rr := httptest.NewRecorder()

	h.FTPFromRamp(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rr.Code)
	}
}

// --- Workouts ---

func TestCreateWorkout(t *testing.T) {
	h, userRepo, _, _ := newTestHandler()

	userRepo.Create(context.Background(), &models.User{
		ID: "user-1", GoogleID: "g1", Email: "a@b.com", DisplayName: "A", FTP: 200,
	})

	body, _ := json.Marshal(models.CreateWorkoutRequest{
		Name:        "Sweet Spot",
		Description: "2x20",
		Intervals: []models.Interval{
			{Type: models.IntervalSteady, DurationSec: 1200, PowerStart: 0.90, PowerEnd: 0.90},
		},
	})
	req := requestWithUser("POST", "/api/workouts", body, "user-1")
	rr := httptest.NewRecorder()

	h.CreateWorkout(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d", rr.Code)
	}

	var w models.Workout
	json.NewDecoder(rr.Body).Decode(&w)
	if w.Name != "Sweet Spot" {
		t.Errorf("name: got %q, want %q", w.Name, "Sweet Spot")
	}
	if w.ID == "" {
		t.Error("expected non-empty ID")
	}
	if len(w.Intervals) != 1 {
		t.Errorf("intervals length: got %d, want 1", len(w.Intervals))
	}
}

func TestCreateWorkoutDefaultName(t *testing.T) {
	h, userRepo, _, _ := newTestHandler()

	userRepo.Create(context.Background(), &models.User{
		ID: "user-1", GoogleID: "g1", Email: "a@b.com", DisplayName: "A", FTP: 200,
	})

	body, _ := json.Marshal(models.CreateWorkoutRequest{})
	req := requestWithUser("POST", "/api/workouts", body, "user-1")
	rr := httptest.NewRecorder()

	h.CreateWorkout(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d", rr.Code)
	}

	var w models.Workout
	json.NewDecoder(rr.Body).Decode(&w)
	if w.Name != "Untitled Workout" {
		t.Errorf("name: got %q, want %q", w.Name, "Untitled Workout")
	}
}

func TestGetWorkout(t *testing.T) {
	h, userRepo, _, workoutRepo := newTestHandler()

	userRepo.Create(context.Background(), &models.User{
		ID: "user-1", GoogleID: "g1", Email: "a@b.com", DisplayName: "A", FTP: 200,
	})

	workoutRepo.Create(context.Background(), &models.Workout{
		ID:     "w-1",
		UserID: "user-1",
		Name:   "Test",
		Intervals: []models.Interval{
			{Type: models.IntervalSteady, DurationSec: 300, PowerStart: 0.65, PowerEnd: 0.65},
		},
	})

	// chi URL params require a chi router context.
	req := requestWithUser("GET", "/api/workouts/w-1", nil, "user-1")
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "w-1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	rr := httptest.NewRecorder()

	h.GetWorkout(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rr.Code)
	}

	var w models.Workout
	json.NewDecoder(rr.Body).Decode(&w)
	if w.Name != "Test" {
		t.Errorf("name: got %q, want %q", w.Name, "Test")
	}
}

func TestGetWorkoutNotFound(t *testing.T) {
	h, userRepo, _, _ := newTestHandler()

	userRepo.Create(context.Background(), &models.User{
		ID: "user-1", GoogleID: "g1", Email: "a@b.com", DisplayName: "A", FTP: 200,
	})

	req := requestWithUser("GET", "/api/workouts/nonexistent", nil, "user-1")
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "nonexistent")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	rr := httptest.NewRecorder()

	h.GetWorkout(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", rr.Code)
	}
}

func TestUpdateWorkout(t *testing.T) {
	h, userRepo, _, workoutRepo := newTestHandler()

	userRepo.Create(context.Background(), &models.User{
		ID: "user-1", GoogleID: "g1", Email: "a@b.com", DisplayName: "A", FTP: 200,
	})

	workoutRepo.Create(context.Background(), &models.Workout{
		ID:        "w-1",
		UserID:    "user-1",
		Name:      "Original",
		Intervals: []models.Interval{},
	})

	newName := "Updated"
	body, _ := json.Marshal(models.UpdateWorkoutRequest{
		Name: &newName,
		Intervals: []models.Interval{
			{Type: models.IntervalRamp, DurationSec: 600, PowerStart: 0.50, PowerEnd: 1.00},
		},
	})
	req := requestWithUser("PUT", "/api/workouts/w-1", body, "user-1")
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "w-1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	rr := httptest.NewRecorder()

	h.UpdateWorkout(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rr.Code)
	}

	var w models.Workout
	json.NewDecoder(rr.Body).Decode(&w)
	if w.Name != "Updated" {
		t.Errorf("name: got %q, want %q", w.Name, "Updated")
	}
}

func TestUpdateWorkoutNotFound(t *testing.T) {
	h, userRepo, _, _ := newTestHandler()

	userRepo.Create(context.Background(), &models.User{
		ID: "user-1", GoogleID: "g1", Email: "a@b.com", DisplayName: "A", FTP: 200,
	})

	newName := "Updated"
	body, _ := json.Marshal(models.UpdateWorkoutRequest{Name: &newName})
	req := requestWithUser("PUT", "/api/workouts/nonexistent", body, "user-1")
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "nonexistent")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	rr := httptest.NewRecorder()

	h.UpdateWorkout(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", rr.Code)
	}
}

func TestCreateWorkoutNotAuthenticated(t *testing.T) {
	h, _, _, _ := newTestHandler()

	req := httptest.NewRequest("POST", "/api/workouts", nil)
	rr := httptest.NewRecorder()

	h.CreateWorkout(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", rr.Code)
	}
}

func TestCreateWorkoutInvalidBody(t *testing.T) {
	h, _, _, _ := newTestHandler()

	req := requestWithUser("POST", "/api/workouts", []byte("not json"), "user-1")
	rr := httptest.NewRecorder()

	h.CreateWorkout(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rr.Code)
	}
}

func TestUpdateWorkoutInvalidBody(t *testing.T) {
	h, userRepo, _, workoutRepo := newTestHandler()

	userRepo.Create(context.Background(), &models.User{
		ID: "user-1", GoogleID: "g1", Email: "a@b.com", DisplayName: "A", FTP: 200,
	})
	workoutRepo.Create(context.Background(), &models.Workout{
		ID: "w-1", UserID: "user-1", Name: "Test", Intervals: []models.Interval{},
	})

	req := requestWithUser("PUT", "/api/workouts/w-1", []byte("not json"), "user-1")
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "w-1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	rr := httptest.NewRecorder()

	h.UpdateWorkout(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rr.Code)
	}
}

// --- OAuth ---

func TestGoogleLogin(t *testing.T) {
	h, _, _, _ := newTestHandler()

	req := httptest.NewRequest("GET", "/auth/google/login", nil)
	rr := httptest.NewRecorder()

	h.GoogleLogin(rr, req)

	if rr.Code != http.StatusTemporaryRedirect {
		t.Errorf("expected status 307, got %d", rr.Code)
	}

	location := rr.Header().Get("Location")
	if location == "" {
		t.Error("expected Location header")
	}
}

func TestGoogleCallbackMissingCode(t *testing.T) {
	h, _, _, _ := newTestHandler()

	req := httptest.NewRequest("GET", "/auth/google/callback", nil)
	rr := httptest.NewRecorder()

	h.GoogleCallback(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rr.Code)
	}
}

func TestGoogleCallbackExchangeError(t *testing.T) {
	h, _, _, _ := newTestHandler()
	h.OAuthConfig.(*mockOAuthProvider).exchangeErr = errors.New("exchange failed")

	req := httptest.NewRequest("GET", "/auth/google/callback?code=test-code", nil)
	rr := httptest.NewRecorder()

	h.GoogleCallback(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("expected status 500, got %d", rr.Code)
	}
}

func TestGoogleCallbackNewUser(t *testing.T) {
	h, userRepo, _, _ := newTestHandler()

	// Set up a mock user info fetcher.
	h.SetUserInfoFetcher(func(ctx context.Context, client *http.Client) (*GoogleUserInfo, error) {
		return &GoogleUserInfo{
			Sub:   "google-new",
			Email: "new@example.com",
			Name:  "New User",
		}, nil
	})

	req := httptest.NewRequest("GET", "/auth/google/callback?code=test-code", nil)
	rr := httptest.NewRecorder()

	h.GoogleCallback(rr, req)

	if rr.Code != http.StatusTemporaryRedirect {
		t.Fatalf("expected status 307, got %d", rr.Code)
	}

	// Verify user was created.
	u, _ := userRepo.GetByGoogleID(context.Background(), "google-new")
	if u == nil {
		t.Fatal("expected user to be created")
	}
	if u.Email != "new@example.com" {
		t.Errorf("email: got %q, want %q", u.Email, "new@example.com")
	}
	if u.FTP != 200 {
		t.Errorf("default FTP: got %d, want 200", u.FTP)
	}
}

func TestGoogleCallbackExistingUser(t *testing.T) {
	h, userRepo, _, _ := newTestHandler()

	// Pre-create user.
	userRepo.Create(context.Background(), &models.User{
		ID: "existing-user", GoogleID: "google-existing", Email: "old@example.com", DisplayName: "Old", FTP: 250,
	})

	h.SetUserInfoFetcher(func(ctx context.Context, client *http.Client) (*GoogleUserInfo, error) {
		return &GoogleUserInfo{
			Sub:   "google-existing",
			Email: "updated@example.com",
			Name:  "Updated Name",
		}, nil
	})

	req := httptest.NewRequest("GET", "/auth/google/callback?code=test-code", nil)
	rr := httptest.NewRecorder()

	h.GoogleCallback(rr, req)

	if rr.Code != http.StatusTemporaryRedirect {
		t.Fatalf("expected status 307, got %d", rr.Code)
	}

	// Verify user was updated.
	u, _ := userRepo.GetByGoogleID(context.Background(), "google-existing")
	if u.Email != "updated@example.com" {
		t.Errorf("email: got %q, want %q", u.Email, "updated@example.com")
	}
	if u.DisplayName != "Updated Name" {
		t.Errorf("display_name: got %q, want %q", u.DisplayName, "Updated Name")
	}
	// FTP should not change.
	if u.FTP != 250 {
		t.Errorf("FTP should not change: got %d, want 250", u.FTP)
	}
}

// --- Error scenarios with repo failures ---

func TestGetProfileDBError(t *testing.T) {
	h, userRepo, _, _ := newTestHandler()
	userRepo.getErr = errors.New("db error")

	req := requestWithUser("GET", "/api/user/profile", nil, "user-1")
	rr := httptest.NewRecorder()

	h.GetProfile(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("expected status 500, got %d", rr.Code)
	}
}

func TestCreateWorkoutRepoError(t *testing.T) {
	h, _, _, workoutRepo := newTestHandler()
	workoutRepo.createErr = errors.New("create failed")

	body, _ := json.Marshal(models.CreateWorkoutRequest{Name: "Test"})
	req := requestWithUser("POST", "/api/workouts", body, "user-1")
	rr := httptest.NewRecorder()

	h.CreateWorkout(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("expected status 500, got %d", rr.Code)
	}
}

func TestAddFTPHistoryRepoError(t *testing.T) {
	h, _, ftpRepo, _ := newTestHandler()
	ftpRepo.createErr = errors.New("create failed")

	body, _ := json.Marshal(models.AddFTPHistoryRequest{
		FTP: 250, RecordedAt: "2025-06-15",
	})
	req := requestWithUser("POST", "/api/user/ftp-history", body, "user-1")
	rr := httptest.NewRecorder()

	h.AddFTPHistory(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("expected status 500, got %d", rr.Code)
	}
}

// --- Additional error path tests ---

func TestGetFTPHistoryNotAuthenticated(t *testing.T) {
	h, _, _, _ := newTestHandler()

	req := httptest.NewRequest("GET", "/api/user/ftp-history", nil)
	rr := httptest.NewRecorder()

	h.GetFTPHistory(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", rr.Code)
	}
}

func TestGetFTPHistoryDBError(t *testing.T) {
	h, _, ftpRepo, _ := newTestHandler()
	ftpRepo.listErr = errors.New("db error")

	req := requestWithUser("GET", "/api/user/ftp-history", nil, "user-1")
	rr := httptest.NewRecorder()

	h.GetFTPHistory(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("expected status 500, got %d", rr.Code)
	}
}

func TestAddFTPHistoryNotAuthenticated(t *testing.T) {
	h, _, _, _ := newTestHandler()

	req := httptest.NewRequest("POST", "/api/user/ftp-history", nil)
	rr := httptest.NewRecorder()

	h.AddFTPHistory(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", rr.Code)
	}
}

func TestAddFTPHistoryInvalidBody(t *testing.T) {
	h, _, _, _ := newTestHandler()

	req := requestWithUser("POST", "/api/user/ftp-history", []byte("not json"), "user-1")
	rr := httptest.NewRecorder()

	h.AddFTPHistory(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rr.Code)
	}
}

func TestAddFTPHistoryFTPTooHigh(t *testing.T) {
	h, _, _, _ := newTestHandler()

	body, _ := json.Marshal(models.AddFTPHistoryRequest{
		FTP: 1500, RecordedAt: "2025-06-15",
	})
	req := requestWithUser("POST", "/api/user/ftp-history", body, "user-1")
	rr := httptest.NewRecorder()

	h.AddFTPHistory(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rr.Code)
	}
}

func TestFTPFromRampNotAuthenticated(t *testing.T) {
	h, _, _, _ := newTestHandler()

	req := httptest.NewRequest("POST", "/api/user/ftp-from-ramp", nil)
	rr := httptest.NewRecorder()

	h.FTPFromRamp(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", rr.Code)
	}
}

func TestFTPFromRampInvalidBody(t *testing.T) {
	h, _, _, _ := newTestHandler()

	req := requestWithUser("POST", "/api/user/ftp-from-ramp", []byte("not json"), "user-1")
	rr := httptest.NewRecorder()

	h.FTPFromRamp(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rr.Code)
	}
}

func TestFTPFromRampConfirmDBError(t *testing.T) {
	h, userRepo, ftpRepo, _ := newTestHandler()

	userRepo.Create(context.Background(), &models.User{
		ID: "user-1", GoogleID: "g1", Email: "a@b.com", DisplayName: "A", FTP: 200,
	})
	ftpRepo.createErr = errors.New("db error")

	body, _ := json.Marshal(models.FTPFromRampRequest{
		MaxOnMinPower: 400,
		Confirm:       true,
	})
	req := requestWithUser("POST", "/api/user/ftp-from-ramp", body, "user-1")
	rr := httptest.NewRecorder()

	h.FTPFromRamp(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("expected status 500, got %d", rr.Code)
	}
}

func TestFTPFromRampConfirmUserNotFound(t *testing.T) {
	h, _, _, _ := newTestHandler()

	body, _ := json.Marshal(models.FTPFromRampRequest{
		MaxOnMinPower: 400,
		Confirm:       true,
	})
	req := requestWithUser("POST", "/api/user/ftp-from-ramp", body, "user-1")
	rr := httptest.NewRecorder()

	h.FTPFromRamp(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("expected status 500, got %d", rr.Code)
	}
}

func TestUpdateProfileNotAuthenticated(t *testing.T) {
	h, _, _, _ := newTestHandler()

	req := httptest.NewRequest("PUT", "/api/user/profile", nil)
	rr := httptest.NewRecorder()

	h.UpdateProfile(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", rr.Code)
	}
}

func TestUpdateProfileDBError(t *testing.T) {
	h, userRepo, _, _ := newTestHandler()
	userRepo.getErr = errors.New("db error")

	body, _ := json.Marshal(models.ProfileUpdateRequest{FTP: intPtr(250)})
	req := requestWithUser("PUT", "/api/user/profile", body, "user-1")
	rr := httptest.NewRecorder()

	h.UpdateProfile(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("expected status 500, got %d", rr.Code)
	}
}

func TestUpdateProfileUserNotFound(t *testing.T) {
	h, _, _, _ := newTestHandler()

	body, _ := json.Marshal(models.ProfileUpdateRequest{FTP: intPtr(250)})
	req := requestWithUser("PUT", "/api/user/profile", body, "nonexistent")
	rr := httptest.NewRecorder()

	h.UpdateProfile(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", rr.Code)
	}
}

func TestUpdateProfileUpdateError(t *testing.T) {
	h, userRepo, _, _ := newTestHandler()

	userRepo.Create(context.Background(), &models.User{
		ID: "user-1", GoogleID: "g1", Email: "a@b.com", DisplayName: "A", FTP: 200,
	})
	userRepo.updateErr = errors.New("update failed")

	body, _ := json.Marshal(models.ProfileUpdateRequest{FTP: intPtr(250)})
	req := requestWithUser("PUT", "/api/user/profile", body, "user-1")
	rr := httptest.NewRecorder()

	h.UpdateProfile(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("expected status 500, got %d", rr.Code)
	}
}

func TestUpdateProfileFTPTooHigh(t *testing.T) {
	h, userRepo, _, _ := newTestHandler()

	userRepo.Create(context.Background(), &models.User{
		ID: "user-1", GoogleID: "g1", Email: "a@b.com", DisplayName: "A", FTP: 200,
	})

	body, _ := json.Marshal(models.ProfileUpdateRequest{FTP: intPtr(1500)})
	req := requestWithUser("PUT", "/api/user/profile", body, "user-1")
	rr := httptest.NewRecorder()

	h.UpdateProfile(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rr.Code)
	}
}

func TestGetWorkoutNotAuthenticated(t *testing.T) {
	h, _, _, _ := newTestHandler()

	req := httptest.NewRequest("GET", "/api/workouts/w-1", nil)
	rr := httptest.NewRecorder()

	h.GetWorkout(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", rr.Code)
	}
}

func TestGetWorkoutDBError(t *testing.T) {
	h, _, _, workoutRepo := newTestHandler()
	workoutRepo.getErr = errors.New("db error")

	req := requestWithUser("GET", "/api/workouts/w-1", nil, "user-1")
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "w-1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	rr := httptest.NewRecorder()

	h.GetWorkout(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("expected status 500, got %d", rr.Code)
	}
}

func TestUpdateWorkoutNotAuthenticated(t *testing.T) {
	h, _, _, _ := newTestHandler()

	req := httptest.NewRequest("PUT", "/api/workouts/w-1", nil)
	rr := httptest.NewRecorder()

	h.UpdateWorkout(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", rr.Code)
	}
}

func TestUpdateWorkoutDBError(t *testing.T) {
	h, _, _, workoutRepo := newTestHandler()
	workoutRepo.getErr = errors.New("db error")

	newName := "Updated"
	body, _ := json.Marshal(models.UpdateWorkoutRequest{Name: &newName})
	req := requestWithUser("PUT", "/api/workouts/w-1", body, "user-1")
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "w-1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	rr := httptest.NewRecorder()

	h.UpdateWorkout(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("expected status 500, got %d", rr.Code)
	}
}

func TestUpdateWorkoutUpdateError(t *testing.T) {
	h, userRepo, _, workoutRepo := newTestHandler()

	userRepo.Create(context.Background(), &models.User{
		ID: "user-1", GoogleID: "g1", Email: "a@b.com", DisplayName: "A", FTP: 200,
	})
	workoutRepo.Create(context.Background(), &models.Workout{
		ID: "w-1", UserID: "user-1", Name: "Test", Intervals: []models.Interval{},
	})
	workoutRepo.updateErr = errors.New("update failed")

	newName := "Updated"
	body, _ := json.Marshal(models.UpdateWorkoutRequest{Name: &newName})
	req := requestWithUser("PUT", "/api/workouts/w-1", body, "user-1")
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "w-1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	rr := httptest.NewRecorder()

	h.UpdateWorkout(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("expected status 500, got %d", rr.Code)
	}
}

func TestGoogleCallbackUserInfoError(t *testing.T) {
	h, _, _, _ := newTestHandler()

	h.SetUserInfoFetcher(func(ctx context.Context, client *http.Client) (*GoogleUserInfo, error) {
		return nil, errors.New("user info failed")
	})

	req := httptest.NewRequest("GET", "/auth/google/callback?code=test-code", nil)
	rr := httptest.NewRecorder()

	h.GoogleCallback(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("expected status 500, got %d", rr.Code)
	}
}

func TestGoogleCallbackDBErrorOnLookup(t *testing.T) {
	h, userRepo, _, _ := newTestHandler()
	userRepo.getErr = errors.New("db error")

	h.SetUserInfoFetcher(func(ctx context.Context, client *http.Client) (*GoogleUserInfo, error) {
		return &GoogleUserInfo{Sub: "g1", Email: "a@b.com", Name: "A"}, nil
	})

	req := httptest.NewRequest("GET", "/auth/google/callback?code=test-code", nil)
	rr := httptest.NewRecorder()

	h.GoogleCallback(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("expected status 500, got %d", rr.Code)
	}
}

func TestGoogleCallbackCreateUserError(t *testing.T) {
	h, userRepo, _, _ := newTestHandler()
	userRepo.createErr = errors.New("create failed")

	h.SetUserInfoFetcher(func(ctx context.Context, client *http.Client) (*GoogleUserInfo, error) {
		return &GoogleUserInfo{Sub: "g-new", Email: "new@b.com", Name: "New"}, nil
	})

	req := httptest.NewRequest("GET", "/auth/google/callback?code=test-code", nil)
	rr := httptest.NewRecorder()

	h.GoogleCallback(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("expected status 500, got %d", rr.Code)
	}
}

func TestGoogleCallbackUpdateUserError(t *testing.T) {
	h, userRepo, _, _ := newTestHandler()

	userRepo.Create(context.Background(), &models.User{
		ID: "u1", GoogleID: "g1", Email: "a@b.com", DisplayName: "A", FTP: 200,
	})
	userRepo.updateErr = errors.New("update failed")

	h.SetUserInfoFetcher(func(ctx context.Context, client *http.Client) (*GoogleUserInfo, error) {
		return &GoogleUserInfo{Sub: "g1", Email: "updated@b.com", Name: "Updated"}, nil
	})

	req := httptest.NewRequest("GET", "/auth/google/callback?code=test-code", nil)
	rr := httptest.NewRecorder()

	h.GoogleCallback(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("expected status 500, got %d", rr.Code)
	}
}

// --- Helpers ---

func intPtr(i int) *int {
	return &i
}
