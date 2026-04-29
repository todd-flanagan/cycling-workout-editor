package repository

import (
	"context"
	"database/sql"
	"testing"

	"github.com/tflanaga/cycling-workout-editor/backend/internal/db"
	"github.com/tflanaga/cycling-workout-editor/backend/internal/models"
)

func setupTestDB(t *testing.T) *sql.DB {
	t.Helper()
	database, err := db.Open(":memory:")
	if err != nil {
		t.Fatalf("failed to open test database: %v", err)
	}
	if err := db.RunMigrations(database); err != nil {
		t.Fatalf("failed to run migrations: %v", err)
	}
	return database
}

// --- User Repository Tests ---

func TestUserCreate(t *testing.T) {
	database := setupTestDB(t)
	defer database.Close()

	repo := NewSQLiteUserRepo(database)
	ctx := context.Background()

	u := &models.User{
		ID:          "user-1",
		GoogleID:    "google-1",
		Email:       "test@example.com",
		DisplayName: "Test User",
		FTP:         250,
	}

	if err := repo.Create(ctx, u); err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	if u.CreatedAt.IsZero() {
		t.Error("CreatedAt should be set after Create")
	}
	if u.UpdatedAt.IsZero() {
		t.Error("UpdatedAt should be set after Create")
	}
}

func TestUserGetByID(t *testing.T) {
	database := setupTestDB(t)
	defer database.Close()

	repo := NewSQLiteUserRepo(database)
	ctx := context.Background()

	u := &models.User{
		ID:          "user-1",
		GoogleID:    "google-1",
		Email:       "test@example.com",
		DisplayName: "Test User",
		FTP:         250,
	}
	if err := repo.Create(ctx, u); err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	found, err := repo.GetByID(ctx, "user-1")
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}
	if found == nil {
		t.Fatal("expected user, got nil")
	}
	if found.Email != "test@example.com" {
		t.Errorf("Email: got %q, want %q", found.Email, "test@example.com")
	}
	if found.FTP != 250 {
		t.Errorf("FTP: got %d, want 250", found.FTP)
	}
}

func TestUserGetByIDNotFound(t *testing.T) {
	database := setupTestDB(t)
	defer database.Close()

	repo := NewSQLiteUserRepo(database)
	ctx := context.Background()

	found, err := repo.GetByID(ctx, "nonexistent")
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}
	if found != nil {
		t.Error("expected nil for nonexistent user")
	}
}

func TestUserGetByGoogleID(t *testing.T) {
	database := setupTestDB(t)
	defer database.Close()

	repo := NewSQLiteUserRepo(database)
	ctx := context.Background()

	u := &models.User{
		ID:          "user-1",
		GoogleID:    "google-1",
		Email:       "test@example.com",
		DisplayName: "Test User",
		FTP:         200,
	}
	if err := repo.Create(ctx, u); err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	found, err := repo.GetByGoogleID(ctx, "google-1")
	if err != nil {
		t.Fatalf("GetByGoogleID failed: %v", err)
	}
	if found == nil {
		t.Fatal("expected user, got nil")
	}
	if found.ID != "user-1" {
		t.Errorf("ID: got %q, want %q", found.ID, "user-1")
	}
}

func TestUserGetByGoogleIDNotFound(t *testing.T) {
	database := setupTestDB(t)
	defer database.Close()

	repo := NewSQLiteUserRepo(database)
	ctx := context.Background()

	found, err := repo.GetByGoogleID(ctx, "nonexistent")
	if err != nil {
		t.Fatalf("GetByGoogleID failed: %v", err)
	}
	if found != nil {
		t.Error("expected nil for nonexistent google ID")
	}
}

func TestUserUpdate(t *testing.T) {
	database := setupTestDB(t)
	defer database.Close()

	repo := NewSQLiteUserRepo(database)
	ctx := context.Background()

	u := &models.User{
		ID:          "user-1",
		GoogleID:    "google-1",
		Email:       "test@example.com",
		DisplayName: "Test User",
		FTP:         200,
	}
	if err := repo.Create(ctx, u); err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	u.FTP = 275
	u.DisplayName = "Updated Name"
	if err := repo.Update(ctx, u); err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	found, err := repo.GetByID(ctx, "user-1")
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}
	if found.FTP != 275 {
		t.Errorf("FTP: got %d, want 275", found.FTP)
	}
	if found.DisplayName != "Updated Name" {
		t.Errorf("DisplayName: got %q, want %q", found.DisplayName, "Updated Name")
	}
}

func TestUserCreateDuplicateGoogleID(t *testing.T) {
	database := setupTestDB(t)
	defer database.Close()

	repo := NewSQLiteUserRepo(database)
	ctx := context.Background()

	u1 := &models.User{ID: "u1", GoogleID: "g1", Email: "a@b.com", DisplayName: "A", FTP: 200}
	if err := repo.Create(ctx, u1); err != nil {
		t.Fatalf("Create u1 failed: %v", err)
	}

	u2 := &models.User{ID: "u2", GoogleID: "g1", Email: "c@d.com", DisplayName: "B", FTP: 200}
	err := repo.Create(ctx, u2)
	if err == nil {
		t.Error("expected error for duplicate google_id, got nil")
	}
}

// --- FTP History Repository Tests ---

func createTestUser(t *testing.T, database *sql.DB) {
	t.Helper()
	repo := NewSQLiteUserRepo(database)
	u := &models.User{
		ID:          "user-1",
		GoogleID:    "google-1",
		Email:       "test@example.com",
		DisplayName: "Test",
		FTP:         200,
	}
	if err := repo.Create(context.Background(), u); err != nil {
		t.Fatalf("failed to create test user: %v", err)
	}
}

func TestFTPHistoryCreate(t *testing.T) {
	database := setupTestDB(t)
	defer database.Close()
	createTestUser(t, database)

	repo := NewSQLiteFTPHistoryRepo(database)
	ctx := context.Background()

	entry := &models.FTPHistory{
		ID:         "ftp-1",
		UserID:     "user-1",
		FTP:        260,
		RecordedAt: "2025-06-15",
		Source:     "ramp test",
	}

	if err := repo.Create(ctx, entry); err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if entry.CreatedAt.IsZero() {
		t.Error("CreatedAt should be set")
	}
}

func TestFTPHistoryList(t *testing.T) {
	database := setupTestDB(t)
	defer database.Close()
	createTestUser(t, database)

	repo := NewSQLiteFTPHistoryRepo(database)
	ctx := context.Background()

	entries := []models.FTPHistory{
		{ID: "ftp-1", UserID: "user-1", FTP: 200, RecordedAt: "2025-01-01", Source: "initial"},
		{ID: "ftp-2", UserID: "user-1", FTP: 225, RecordedAt: "2025-03-01", Source: "ramp test"},
		{ID: "ftp-3", UserID: "user-1", FTP: 250, RecordedAt: "2025-06-01"},
	}
	for i := range entries {
		if err := repo.Create(ctx, &entries[i]); err != nil {
			t.Fatalf("Create entry %d failed: %v", i, err)
		}
	}

	list, err := repo.List(ctx, "user-1")
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(list) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(list))
	}

	// Should be ordered by recorded_at DESC.
	if list[0].FTP != 250 {
		t.Errorf("first entry FTP: got %d, want 250", list[0].FTP)
	}
	if list[2].FTP != 200 {
		t.Errorf("last entry FTP: got %d, want 200", list[2].FTP)
	}
}

func TestFTPHistoryListEmpty(t *testing.T) {
	database := setupTestDB(t)
	defer database.Close()
	createTestUser(t, database)

	repo := NewSQLiteFTPHistoryRepo(database)
	ctx := context.Background()

	list, err := repo.List(ctx, "user-1")
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if list != nil {
		t.Errorf("expected nil for empty list, got %v", list)
	}
}

func TestFTPHistoryListDifferentUsers(t *testing.T) {
	database := setupTestDB(t)
	defer database.Close()
	createTestUser(t, database)

	// Create second user.
	uRepo := NewSQLiteUserRepo(database)
	u2 := &models.User{ID: "user-2", GoogleID: "google-2", Email: "b@b.com", DisplayName: "B", FTP: 200}
	if err := uRepo.Create(context.Background(), u2); err != nil {
		t.Fatalf("failed to create user-2: %v", err)
	}

	repo := NewSQLiteFTPHistoryRepo(database)
	ctx := context.Background()

	e1 := &models.FTPHistory{ID: "ftp-1", UserID: "user-1", FTP: 250, RecordedAt: "2025-06-01"}
	e2 := &models.FTPHistory{ID: "ftp-2", UserID: "user-2", FTP: 300, RecordedAt: "2025-06-01"}
	repo.Create(ctx, e1)
	repo.Create(ctx, e2)

	list, err := repo.List(ctx, "user-1")
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(list) != 1 {
		t.Fatalf("expected 1 entry for user-1, got %d", len(list))
	}
	if list[0].FTP != 250 {
		t.Errorf("FTP: got %d, want 250", list[0].FTP)
	}
}

// --- Workout Repository Tests ---

func TestWorkoutCreate(t *testing.T) {
	database := setupTestDB(t)
	defer database.Close()
	createTestUser(t, database)

	repo := NewSQLiteWorkoutRepo(database)
	ctx := context.Background()

	w := &models.Workout{
		ID:          "w-1",
		UserID:      "user-1",
		Name:        "Test Workout",
		Description: "A test workout",
		Intervals: []models.Interval{
			{Type: models.IntervalSteady, DurationSec: 600, PowerStart: 0.65, PowerEnd: 0.65},
			{Type: models.IntervalRamp, DurationSec: 300, PowerStart: 0.50, PowerEnd: 1.00},
		},
	}

	if err := repo.Create(ctx, w); err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if w.CreatedAt.IsZero() {
		t.Error("CreatedAt should be set")
	}
}

func TestWorkoutGetByID(t *testing.T) {
	database := setupTestDB(t)
	defer database.Close()
	createTestUser(t, database)

	repo := NewSQLiteWorkoutRepo(database)
	ctx := context.Background()

	w := &models.Workout{
		ID:          "w-1",
		UserID:      "user-1",
		Name:        "Test Workout",
		Description: "A test workout",
		Intervals: []models.Interval{
			{Type: models.IntervalSteady, DurationSec: 600, PowerStart: 0.65, PowerEnd: 0.65},
		},
	}
	if err := repo.Create(ctx, w); err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	found, err := repo.GetByID(ctx, "w-1", "user-1")
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}
	if found == nil {
		t.Fatal("expected workout, got nil")
	}
	if found.Name != "Test Workout" {
		t.Errorf("Name: got %q, want %q", found.Name, "Test Workout")
	}
	if len(found.Intervals) != 1 {
		t.Fatalf("Intervals length: got %d, want 1", len(found.Intervals))
	}
	if found.Intervals[0].Type != models.IntervalSteady {
		t.Errorf("interval type: got %q, want %q", found.Intervals[0].Type, models.IntervalSteady)
	}
}

func TestWorkoutGetByIDNotFound(t *testing.T) {
	database := setupTestDB(t)
	defer database.Close()
	createTestUser(t, database)

	repo := NewSQLiteWorkoutRepo(database)
	ctx := context.Background()

	found, err := repo.GetByID(ctx, "nonexistent", "user-1")
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}
	if found != nil {
		t.Error("expected nil for nonexistent workout")
	}
}

func TestWorkoutGetByIDWrongUser(t *testing.T) {
	database := setupTestDB(t)
	defer database.Close()
	createTestUser(t, database)

	repo := NewSQLiteWorkoutRepo(database)
	ctx := context.Background()

	w := &models.Workout{
		ID:        "w-1",
		UserID:    "user-1",
		Name:      "Test",
		Intervals: []models.Interval{},
	}
	if err := repo.Create(ctx, w); err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// Try to get it as a different user.
	found, err := repo.GetByID(ctx, "w-1", "user-999")
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}
	if found != nil {
		t.Error("expected nil when querying wrong user_id")
	}
}

func TestWorkoutUpdate(t *testing.T) {
	database := setupTestDB(t)
	defer database.Close()
	createTestUser(t, database)

	repo := NewSQLiteWorkoutRepo(database)
	ctx := context.Background()

	w := &models.Workout{
		ID:          "w-1",
		UserID:      "user-1",
		Name:        "Original",
		Description: "Original desc",
		Intervals: []models.Interval{
			{Type: models.IntervalSteady, DurationSec: 300, PowerStart: 0.5, PowerEnd: 0.5},
		},
	}
	if err := repo.Create(ctx, w); err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	w.Name = "Updated"
	w.Description = "Updated desc"
	w.Intervals = []models.Interval{
		{Type: models.IntervalRamp, DurationSec: 600, PowerStart: 0.50, PowerEnd: 1.00},
		{Type: models.IntervalRest, DurationSec: 120, PowerStart: 0.40, PowerEnd: 0.40},
	}

	if err := repo.Update(ctx, w); err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	found, err := repo.GetByID(ctx, "w-1", "user-1")
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}
	if found.Name != "Updated" {
		t.Errorf("Name: got %q, want %q", found.Name, "Updated")
	}
	if len(found.Intervals) != 2 {
		t.Fatalf("Intervals length: got %d, want 2", len(found.Intervals))
	}
}

func TestWorkoutWithTemplateID(t *testing.T) {
	database := setupTestDB(t)
	defer database.Close()
	createTestUser(t, database)

	repo := NewSQLiteWorkoutRepo(database)
	ctx := context.Background()

	tmplID := "tmpl-1"
	w := &models.Workout{
		ID:         "w-1",
		UserID:     "user-1",
		Name:       "From Template",
		Intervals:  []models.Interval{},
		TemplateID: &tmplID,
	}
	if err := repo.Create(ctx, w); err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	found, err := repo.GetByID(ctx, "w-1", "user-1")
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}
	if found.TemplateID == nil {
		t.Fatal("expected TemplateID to be set")
	}
	if *found.TemplateID != tmplID {
		t.Errorf("TemplateID: got %q, want %q", *found.TemplateID, tmplID)
	}
}

func TestWorkoutWithoutTemplateID(t *testing.T) {
	database := setupTestDB(t)
	defer database.Close()
	createTestUser(t, database)

	repo := NewSQLiteWorkoutRepo(database)
	ctx := context.Background()

	w := &models.Workout{
		ID:        "w-1",
		UserID:    "user-1",
		Name:      "No Template",
		Intervals: []models.Interval{},
	}
	if err := repo.Create(ctx, w); err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	found, err := repo.GetByID(ctx, "w-1", "user-1")
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}
	if found.TemplateID != nil {
		t.Errorf("expected TemplateID to be nil, got %v", found.TemplateID)
	}
}
