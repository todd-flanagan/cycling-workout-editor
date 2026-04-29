package models

import (
	"encoding/json"
	"time"
)

// User represents a registered user.
type User struct {
	ID          string    `json:"id"`
	GoogleID    string    `json:"google_id"`
	Email       string    `json:"email"`
	DisplayName string    `json:"display_name"`
	FTP         int       `json:"ftp"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// FTPHistory represents a single FTP measurement record.
type FTPHistory struct {
	ID         string    `json:"id"`
	UserID     string    `json:"user_id"`
	FTP        int       `json:"ftp"`
	RecordedAt string    `json:"recorded_at"` // DATE string YYYY-MM-DD
	Source     string    `json:"source,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
}

// IntervalType defines the kind of interval.
type IntervalType string

const (
	IntervalSteady   IntervalType = "steady"
	IntervalRamp     IntervalType = "ramp"
	IntervalFreeRide IntervalType = "free_ride"
	IntervalRest     IntervalType = "rest"
)

// Interval represents a single interval within a workout.
type Interval struct {
	Type        IntervalType `json:"type"`
	DurationSec int          `json:"duration_sec"`
	PowerStart  float64      `json:"power_start"`
	PowerEnd    float64      `json:"power_end"`
	Cadence     *int         `json:"cadence,omitempty"`
}

// Workout represents a user's workout.
type Workout struct {
	ID          string     `json:"id"`
	UserID      string     `json:"user_id"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
	Intervals   []Interval `json:"intervals"`
	TemplateID  *string    `json:"template_id,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// IntervalsJSON returns the intervals field serialized as a JSON string,
// suitable for storing in the database.
func (w *Workout) IntervalsJSON() (string, error) {
	b, err := json.Marshal(w.Intervals)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// ParseIntervals deserializes a JSON string into the Intervals field.
func (w *Workout) ParseIntervals(data string) error {
	return json.Unmarshal([]byte(data), &w.Intervals)
}

// FTPFromRampRequest is the request body for calculating FTP from a ramp test.
type FTPFromRampRequest struct {
	MaxOnMinPower int  `json:"max_one_min_power"`
	Confirm       bool `json:"confirm"`
}

// FTPFromRampResponse is the response for the FTP from ramp calculation.
type FTPFromRampResponse struct {
	EstimatedFTP int  `json:"estimated_ftp"`
	Confirmed    bool `json:"confirmed"`
}

// ProfileUpdateRequest is the request body for updating a user's profile.
type ProfileUpdateRequest struct {
	DisplayName *string `json:"display_name,omitempty"`
	FTP         *int    `json:"ftp,omitempty"`
}

// CreateWorkoutRequest is the request body for creating a workout.
type CreateWorkoutRequest struct {
	Name        string     `json:"name"`
	Description string     `json:"description"`
	Intervals   []Interval `json:"intervals"`
}

// UpdateWorkoutRequest is the request body for updating a workout.
type UpdateWorkoutRequest struct {
	Name        *string    `json:"name,omitempty"`
	Description *string    `json:"description,omitempty"`
	Intervals   []Interval `json:"intervals,omitempty"`
}

// AddFTPHistoryRequest is the request body for adding an FTP history entry.
type AddFTPHistoryRequest struct {
	FTP        int    `json:"ftp"`
	RecordedAt string `json:"recorded_at"` // YYYY-MM-DD
	Source     string `json:"source,omitempty"`
}

// ErrorResponse is the standard error response format.
type ErrorResponse struct {
	Error string `json:"error"`
}
