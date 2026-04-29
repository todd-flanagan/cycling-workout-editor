package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/tflanaga/cycling-workout-editor/backend/internal/models"
)

// Ensure compile-time interface compliance.
var (
	_ UserRepository       = (*SQLiteUserRepo)(nil)
	_ FTPHistoryRepository = (*SQLiteFTPHistoryRepo)(nil)
	_ WorkoutRepository    = (*SQLiteWorkoutRepo)(nil)
)

// SQLiteUserRepo implements UserRepository using SQLite.
type SQLiteUserRepo struct {
	db *sql.DB
}

// NewSQLiteUserRepo creates a new SQLiteUserRepo.
func NewSQLiteUserRepo(db *sql.DB) *SQLiteUserRepo {
	return &SQLiteUserRepo{db: db}
}

func (r *SQLiteUserRepo) GetByID(ctx context.Context, id string) (*models.User, error) {
	u := &models.User{}
	err := r.db.QueryRowContext(ctx,
		`SELECT id, google_id, email, display_name, ftp, created_at, updated_at
		 FROM users WHERE id = ?`, id,
	).Scan(&u.ID, &u.GoogleID, &u.Email, &u.DisplayName, &u.FTP, &u.CreatedAt, &u.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get user by id: %w", err)
	}
	return u, nil
}

func (r *SQLiteUserRepo) GetByGoogleID(ctx context.Context, googleID string) (*models.User, error) {
	u := &models.User{}
	err := r.db.QueryRowContext(ctx,
		`SELECT id, google_id, email, display_name, ftp, created_at, updated_at
		 FROM users WHERE google_id = ?`, googleID,
	).Scan(&u.ID, &u.GoogleID, &u.Email, &u.DisplayName, &u.FTP, &u.CreatedAt, &u.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get user by google_id: %w", err)
	}
	return u, nil
}

func (r *SQLiteUserRepo) Create(ctx context.Context, user *models.User) error {
	now := time.Now().UTC()
	user.CreatedAt = now
	user.UpdatedAt = now
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO users (id, google_id, email, display_name, ftp, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		user.ID, user.GoogleID, user.Email, user.DisplayName, user.FTP, user.CreatedAt, user.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("create user: %w", err)
	}
	return nil
}

func (r *SQLiteUserRepo) Update(ctx context.Context, user *models.User) error {
	user.UpdatedAt = time.Now().UTC()
	_, err := r.db.ExecContext(ctx,
		`UPDATE users SET email = ?, display_name = ?, ftp = ?, updated_at = ? WHERE id = ?`,
		user.Email, user.DisplayName, user.FTP, user.UpdatedAt, user.ID,
	)
	if err != nil {
		return fmt.Errorf("update user: %w", err)
	}
	return nil
}

// SQLiteFTPHistoryRepo implements FTPHistoryRepository using SQLite.
type SQLiteFTPHistoryRepo struct {
	db *sql.DB
}

// NewSQLiteFTPHistoryRepo creates a new SQLiteFTPHistoryRepo.
func NewSQLiteFTPHistoryRepo(db *sql.DB) *SQLiteFTPHistoryRepo {
	return &SQLiteFTPHistoryRepo{db: db}
}

func (r *SQLiteFTPHistoryRepo) List(ctx context.Context, userID string) ([]models.FTPHistory, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, user_id, ftp, recorded_at, source, created_at
		 FROM ftp_history WHERE user_id = ? ORDER BY recorded_at DESC`, userID,
	)
	if err != nil {
		return nil, fmt.Errorf("list ftp_history: %w", err)
	}
	defer rows.Close()

	var entries []models.FTPHistory
	for rows.Next() {
		var e models.FTPHistory
		var source sql.NullString
		if err := rows.Scan(&e.ID, &e.UserID, &e.FTP, &e.RecordedAt, &source, &e.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan ftp_history row: %w", err)
		}
		if source.Valid {
			e.Source = source.String
		}
		entries = append(entries, e)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate ftp_history rows: %w", err)
	}
	return entries, nil
}

func (r *SQLiteFTPHistoryRepo) Create(ctx context.Context, entry *models.FTPHistory) error {
	entry.CreatedAt = time.Now().UTC()
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO ftp_history (id, user_id, ftp, recorded_at, source, created_at)
		 VALUES (?, ?, ?, ?, ?, ?)`,
		entry.ID, entry.UserID, entry.FTP, entry.RecordedAt, entry.Source, entry.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("create ftp_history: %w", err)
	}
	return nil
}

// SQLiteWorkoutRepo implements WorkoutRepository using SQLite.
type SQLiteWorkoutRepo struct {
	db *sql.DB
}

// NewSQLiteWorkoutRepo creates a new SQLiteWorkoutRepo.
func NewSQLiteWorkoutRepo(db *sql.DB) *SQLiteWorkoutRepo {
	return &SQLiteWorkoutRepo{db: db}
}

func (r *SQLiteWorkoutRepo) GetByID(ctx context.Context, id string, userID string) (*models.Workout, error) {
	w := &models.Workout{}
	var intervalsJSON string
	var templateID sql.NullString
	err := r.db.QueryRowContext(ctx,
		`SELECT id, user_id, name, description, intervals, template_id, created_at, updated_at
		 FROM workouts WHERE id = ? AND user_id = ?`, id, userID,
	).Scan(&w.ID, &w.UserID, &w.Name, &w.Description, &intervalsJSON, &templateID, &w.CreatedAt, &w.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get workout by id: %w", err)
	}

	if templateID.Valid {
		w.TemplateID = &templateID.String
	}

	if err := json.Unmarshal([]byte(intervalsJSON), &w.Intervals); err != nil {
		return nil, fmt.Errorf("parse workout intervals: %w", err)
	}

	return w, nil
}

func (r *SQLiteWorkoutRepo) Create(ctx context.Context, workout *models.Workout) error {
	now := time.Now().UTC()
	workout.CreatedAt = now
	workout.UpdatedAt = now

	intervalsJSON, err := json.Marshal(workout.Intervals)
	if err != nil {
		return fmt.Errorf("marshal intervals: %w", err)
	}

	var templateID *string
	if workout.TemplateID != nil {
		templateID = workout.TemplateID
	}

	_, err = r.db.ExecContext(ctx,
		`INSERT INTO workouts (id, user_id, name, description, intervals, template_id, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		workout.ID, workout.UserID, workout.Name, workout.Description, string(intervalsJSON),
		templateID, workout.CreatedAt, workout.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("create workout: %w", err)
	}
	return nil
}

func (r *SQLiteWorkoutRepo) Update(ctx context.Context, workout *models.Workout) error {
	workout.UpdatedAt = time.Now().UTC()

	intervalsJSON, err := json.Marshal(workout.Intervals)
	if err != nil {
		return fmt.Errorf("marshal intervals: %w", err)
	}

	_, err = r.db.ExecContext(ctx,
		`UPDATE workouts SET name = ?, description = ?, intervals = ?, updated_at = ?
		 WHERE id = ? AND user_id = ?`,
		workout.Name, workout.Description, string(intervalsJSON),
		workout.UpdatedAt, workout.ID, workout.UserID,
	)
	if err != nil {
		return fmt.Errorf("update workout: %w", err)
	}
	return nil
}
