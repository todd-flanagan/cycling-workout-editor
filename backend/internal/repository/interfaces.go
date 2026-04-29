package repository

import (
	"context"

	"github.com/tflanaga/cycling-workout-editor/backend/internal/models"
)

// UserRepository defines operations for user persistence.
type UserRepository interface {
	GetByID(ctx context.Context, id string) (*models.User, error)
	GetByGoogleID(ctx context.Context, googleID string) (*models.User, error)
	Create(ctx context.Context, user *models.User) error
	Update(ctx context.Context, user *models.User) error
}

// FTPHistoryRepository defines operations for FTP history persistence.
type FTPHistoryRepository interface {
	List(ctx context.Context, userID string) ([]models.FTPHistory, error)
	Create(ctx context.Context, entry *models.FTPHistory) error
}

// WorkoutRepository defines operations for workout persistence.
type WorkoutRepository interface {
	GetByID(ctx context.Context, id string, userID string) (*models.Workout, error)
	Create(ctx context.Context, workout *models.Workout) error
	Update(ctx context.Context, workout *models.Workout) error
}
