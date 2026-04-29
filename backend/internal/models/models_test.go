package models

import (
	"encoding/json"
	"testing"
	"time"
)

func TestUserJSONMarshal(t *testing.T) {
	u := User{
		ID:          "user-123",
		GoogleID:    "google-456",
		Email:       "test@example.com",
		DisplayName: "Test User",
		FTP:         250,
		CreatedAt:   time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		UpdatedAt:   time.Date(2025, 1, 2, 0, 0, 0, 0, time.UTC),
	}

	data, err := json.Marshal(u)
	if err != nil {
		t.Fatalf("failed to marshal User: %v", err)
	}

	var decoded User
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal User: %v", err)
	}

	if decoded.ID != u.ID {
		t.Errorf("ID mismatch: got %q, want %q", decoded.ID, u.ID)
	}
	if decoded.GoogleID != u.GoogleID {
		t.Errorf("GoogleID mismatch: got %q, want %q", decoded.GoogleID, u.GoogleID)
	}
	if decoded.Email != u.Email {
		t.Errorf("Email mismatch: got %q, want %q", decoded.Email, u.Email)
	}
	if decoded.DisplayName != u.DisplayName {
		t.Errorf("DisplayName mismatch: got %q, want %q", decoded.DisplayName, u.DisplayName)
	}
	if decoded.FTP != u.FTP {
		t.Errorf("FTP mismatch: got %d, want %d", decoded.FTP, u.FTP)
	}
}

func TestUserJSONUnmarshal(t *testing.T) {
	raw := `{"id":"u1","google_id":"g1","email":"a@b.com","display_name":"Alice","ftp":300,"created_at":"2025-06-01T00:00:00Z","updated_at":"2025-06-01T00:00:00Z"}`

	var u User
	if err := json.Unmarshal([]byte(raw), &u); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
	if u.ID != "u1" {
		t.Errorf("ID: got %q, want %q", u.ID, "u1")
	}
	if u.FTP != 300 {
		t.Errorf("FTP: got %d, want %d", u.FTP, 300)
	}
}

func TestFTPHistoryJSON(t *testing.T) {
	h := FTPHistory{
		ID:         "ftp-1",
		UserID:     "user-1",
		FTP:        275,
		RecordedAt: "2025-03-15",
		Source:     "ramp test",
		CreatedAt:  time.Date(2025, 3, 15, 12, 0, 0, 0, time.UTC),
	}

	data, err := json.Marshal(h)
	if err != nil {
		t.Fatalf("failed to marshal FTPHistory: %v", err)
	}

	var decoded FTPHistory
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal FTPHistory: %v", err)
	}

	if decoded.FTP != h.FTP {
		t.Errorf("FTP mismatch: got %d, want %d", decoded.FTP, h.FTP)
	}
	if decoded.Source != h.Source {
		t.Errorf("Source mismatch: got %q, want %q", decoded.Source, h.Source)
	}
	if decoded.RecordedAt != h.RecordedAt {
		t.Errorf("RecordedAt mismatch: got %q, want %q", decoded.RecordedAt, h.RecordedAt)
	}
}

func TestFTPHistoryJSONOmitEmptySource(t *testing.T) {
	h := FTPHistory{
		ID:         "ftp-2",
		UserID:     "user-1",
		FTP:        250,
		RecordedAt: "2025-01-01",
	}

	data, err := json.Marshal(h)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	raw := string(data)
	// Source should be omitted when empty
	var m map[string]interface{}
	if err := json.Unmarshal([]byte(raw), &m); err != nil {
		t.Fatalf("failed to unmarshal to map: %v", err)
	}
	if _, ok := m["source"]; ok {
		t.Error("expected 'source' to be omitted when empty")
	}
}

func TestIntervalJSON(t *testing.T) {
	cadence := 90
	intervals := []Interval{
		{Type: IntervalSteady, DurationSec: 300, PowerStart: 0.75, PowerEnd: 0.75, Cadence: &cadence},
		{Type: IntervalRamp, DurationSec: 600, PowerStart: 0.50, PowerEnd: 1.00},
		{Type: IntervalFreeRide, DurationSec: 120, PowerStart: 0, PowerEnd: 0},
		{Type: IntervalRest, DurationSec: 180, PowerStart: 0.40, PowerEnd: 0.40},
	}

	data, err := json.Marshal(intervals)
	if err != nil {
		t.Fatalf("failed to marshal intervals: %v", err)
	}

	var decoded []Interval
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal intervals: %v", err)
	}

	if len(decoded) != len(intervals) {
		t.Fatalf("length mismatch: got %d, want %d", len(decoded), len(intervals))
	}

	if decoded[0].Type != IntervalSteady {
		t.Errorf("type mismatch: got %q, want %q", decoded[0].Type, IntervalSteady)
	}
	if decoded[0].Cadence == nil || *decoded[0].Cadence != 90 {
		t.Error("cadence should be 90 for the first interval")
	}
	if decoded[1].Cadence != nil {
		t.Error("cadence should be nil for the second interval")
	}
	if decoded[1].Type != IntervalRamp {
		t.Errorf("type mismatch: got %q, want %q", decoded[1].Type, IntervalRamp)
	}
}

func TestWorkoutJSON(t *testing.T) {
	tmplID := "tmpl-1"
	w := Workout{
		ID:          "w-1",
		UserID:      "u-1",
		Name:        "Test Workout",
		Description: "A test",
		Intervals: []Interval{
			{Type: IntervalSteady, DurationSec: 600, PowerStart: 0.65, PowerEnd: 0.65},
		},
		TemplateID: &tmplID,
		CreatedAt:  time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		UpdatedAt:  time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
	}

	data, err := json.Marshal(w)
	if err != nil {
		t.Fatalf("failed to marshal Workout: %v", err)
	}

	var decoded Workout
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal Workout: %v", err)
	}

	if decoded.Name != w.Name {
		t.Errorf("Name mismatch: got %q, want %q", decoded.Name, w.Name)
	}
	if len(decoded.Intervals) != 1 {
		t.Fatalf("Intervals length: got %d, want 1", len(decoded.Intervals))
	}
	if decoded.TemplateID == nil || *decoded.TemplateID != tmplID {
		t.Error("TemplateID should be set")
	}
}

func TestWorkoutJSONNoTemplateID(t *testing.T) {
	w := Workout{
		ID:        "w-2",
		UserID:    "u-1",
		Name:      "Plain Workout",
		Intervals: []Interval{},
	}

	data, err := json.Marshal(w)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var m map[string]interface{}
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("failed to unmarshal to map: %v", err)
	}
	if _, ok := m["template_id"]; ok {
		t.Error("template_id should be omitted when nil")
	}
}

func TestWorkoutIntervalsJSON(t *testing.T) {
	w := Workout{
		Intervals: []Interval{
			{Type: IntervalSteady, DurationSec: 300, PowerStart: 0.75, PowerEnd: 0.75},
			{Type: IntervalRamp, DurationSec: 600, PowerStart: 0.50, PowerEnd: 1.00},
		},
	}

	jsonStr, err := w.IntervalsJSON()
	if err != nil {
		t.Fatalf("IntervalsJSON failed: %v", err)
	}

	w2 := Workout{}
	if err := w2.ParseIntervals(jsonStr); err != nil {
		t.Fatalf("ParseIntervals failed: %v", err)
	}

	if len(w2.Intervals) != 2 {
		t.Fatalf("Intervals length: got %d, want 2", len(w2.Intervals))
	}
	if w2.Intervals[0].Type != IntervalSteady {
		t.Errorf("first interval type: got %q, want %q", w2.Intervals[0].Type, IntervalSteady)
	}
	if w2.Intervals[1].PowerEnd != 1.0 {
		t.Errorf("second interval power_end: got %f, want 1.0", w2.Intervals[1].PowerEnd)
	}
}

func TestErrorResponseJSON(t *testing.T) {
	e := ErrorResponse{Error: "something went wrong"}
	data, err := json.Marshal(e)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var decoded ErrorResponse
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
	if decoded.Error != e.Error {
		t.Errorf("Error mismatch: got %q, want %q", decoded.Error, e.Error)
	}
}

func TestFTPFromRampRequestJSON(t *testing.T) {
	raw := `{"max_one_min_power":400,"confirm":true}`
	var req FTPFromRampRequest
	if err := json.Unmarshal([]byte(raw), &req); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
	if req.MaxOnMinPower != 400 {
		t.Errorf("MaxOnMinPower: got %d, want 400", req.MaxOnMinPower)
	}
	if !req.Confirm {
		t.Error("Confirm should be true")
	}
}

func TestProfileUpdateRequestJSON(t *testing.T) {
	raw := `{"display_name":"New Name","ftp":280}`
	var req ProfileUpdateRequest
	if err := json.Unmarshal([]byte(raw), &req); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
	if req.DisplayName == nil || *req.DisplayName != "New Name" {
		t.Error("DisplayName should be 'New Name'")
	}
	if req.FTP == nil || *req.FTP != 280 {
		t.Error("FTP should be 280")
	}
}

func TestProfileUpdateRequestPartialJSON(t *testing.T) {
	raw := `{"ftp":300}`
	var req ProfileUpdateRequest
	if err := json.Unmarshal([]byte(raw), &req); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
	if req.DisplayName != nil {
		t.Error("DisplayName should be nil when not provided")
	}
	if req.FTP == nil || *req.FTP != 300 {
		t.Error("FTP should be 300")
	}
}

func TestCreateWorkoutRequestJSON(t *testing.T) {
	raw := `{"name":"My Workout","description":"Test","intervals":[{"type":"steady","duration_sec":300,"power_start":0.75,"power_end":0.75}]}`
	var req CreateWorkoutRequest
	if err := json.Unmarshal([]byte(raw), &req); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
	if req.Name != "My Workout" {
		t.Errorf("Name: got %q, want %q", req.Name, "My Workout")
	}
	if len(req.Intervals) != 1 {
		t.Fatalf("Intervals length: got %d, want 1", len(req.Intervals))
	}
}

func TestAddFTPHistoryRequestJSON(t *testing.T) {
	raw := `{"ftp":260,"recorded_at":"2025-06-15","source":"20-min test"}`
	var req AddFTPHistoryRequest
	if err := json.Unmarshal([]byte(raw), &req); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
	if req.FTP != 260 {
		t.Errorf("FTP: got %d, want 260", req.FTP)
	}
	if req.RecordedAt != "2025-06-15" {
		t.Errorf("RecordedAt: got %q, want %q", req.RecordedAt, "2025-06-15")
	}
	if req.Source != "20-min test" {
		t.Errorf("Source: got %q, want %q", req.Source, "20-min test")
	}
}
