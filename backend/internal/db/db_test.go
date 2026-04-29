package db

import (
	"testing"
)

func TestOpen(t *testing.T) {
	database, err := Open(":memory:")
	if err != nil {
		t.Fatalf("failed to open in-memory database: %v", err)
	}
	defer database.Close()

	// Verify connection works.
	if err := database.Ping(); err != nil {
		t.Fatalf("failed to ping database: %v", err)
	}

	// Verify foreign keys are enabled.
	var fkEnabled int
	if err := database.QueryRow("PRAGMA foreign_keys").Scan(&fkEnabled); err != nil {
		t.Fatalf("failed to check foreign_keys pragma: %v", err)
	}
	if fkEnabled != 1 {
		t.Error("foreign_keys should be enabled")
	}
}

func TestRunMigrations(t *testing.T) {
	database, err := Open(":memory:")
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
	defer database.Close()

	if err := RunMigrations(database); err != nil {
		t.Fatalf("RunMigrations failed: %v", err)
	}

	// Verify schema_migrations table has entries.
	var count int
	if err := database.QueryRow("SELECT COUNT(*) FROM schema_migrations").Scan(&count); err != nil {
		t.Fatalf("failed to query schema_migrations: %v", err)
	}
	if count != len(migrations) {
		t.Errorf("expected %d migrations applied, got %d", len(migrations), count)
	}

	// Verify tables exist by inserting and querying.
	tables := []string{"users", "ftp_history", "workouts"}
	for _, table := range tables {
		var n int
		err := database.QueryRow("SELECT COUNT(*) FROM " + table).Scan(&n)
		if err != nil {
			t.Errorf("table %q should exist but query failed: %v", table, err)
		}
	}
}

func TestRunMigrationsIdempotent(t *testing.T) {
	database, err := Open(":memory:")
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
	defer database.Close()

	// Run migrations twice — should be idempotent.
	if err := RunMigrations(database); err != nil {
		t.Fatalf("first RunMigrations failed: %v", err)
	}
	if err := RunMigrations(database); err != nil {
		t.Fatalf("second RunMigrations failed: %v", err)
	}

	var count int
	if err := database.QueryRow("SELECT COUNT(*) FROM schema_migrations").Scan(&count); err != nil {
		t.Fatalf("failed to query schema_migrations: %v", err)
	}
	if count != len(migrations) {
		t.Errorf("expected %d migrations, got %d after running twice", len(migrations), count)
	}
}

func TestMigrationUsersTable(t *testing.T) {
	database, err := Open(":memory:")
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
	defer database.Close()

	if err := RunMigrations(database); err != nil {
		t.Fatalf("RunMigrations failed: %v", err)
	}

	// Insert a user.
	_, err = database.Exec(
		`INSERT INTO users (id, google_id, email, display_name, ftp) VALUES (?, ?, ?, ?, ?)`,
		"u1", "g1", "test@example.com", "Test User", 250,
	)
	if err != nil {
		t.Fatalf("failed to insert user: %v", err)
	}

	// Query back.
	var email string
	var ftp int
	if err := database.QueryRow("SELECT email, ftp FROM users WHERE id = ?", "u1").Scan(&email, &ftp); err != nil {
		t.Fatalf("failed to query user: %v", err)
	}
	if email != "test@example.com" {
		t.Errorf("email: got %q, want %q", email, "test@example.com")
	}
	if ftp != 250 {
		t.Errorf("ftp: got %d, want 250", ftp)
	}
}

func TestMigrationFTPHistoryTable(t *testing.T) {
	database, err := Open(":memory:")
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
	defer database.Close()

	if err := RunMigrations(database); err != nil {
		t.Fatalf("RunMigrations failed: %v", err)
	}

	// Insert user first (FK constraint).
	_, err = database.Exec(
		`INSERT INTO users (id, google_id, email, display_name) VALUES (?, ?, ?, ?)`,
		"u1", "g1", "test@example.com", "Test",
	)
	if err != nil {
		t.Fatalf("failed to insert user: %v", err)
	}

	// Insert FTP history entry.
	_, err = database.Exec(
		`INSERT INTO ftp_history (id, user_id, ftp, recorded_at, source) VALUES (?, ?, ?, ?, ?)`,
		"ftp1", "u1", 275, "2025-06-15", "ramp test",
	)
	if err != nil {
		t.Fatalf("failed to insert ftp_history: %v", err)
	}

	var ftp int
	var source string
	if err := database.QueryRow("SELECT ftp, source FROM ftp_history WHERE id = ?", "ftp1").Scan(&ftp, &source); err != nil {
		t.Fatalf("failed to query ftp_history: %v", err)
	}
	if ftp != 275 {
		t.Errorf("ftp: got %d, want 275", ftp)
	}
}

func TestMigrationWorkoutsTable(t *testing.T) {
	database, err := Open(":memory:")
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
	defer database.Close()

	if err := RunMigrations(database); err != nil {
		t.Fatalf("RunMigrations failed: %v", err)
	}

	// Insert user first.
	_, err = database.Exec(
		`INSERT INTO users (id, google_id, email, display_name) VALUES (?, ?, ?, ?)`,
		"u1", "g1", "test@example.com", "Test",
	)
	if err != nil {
		t.Fatalf("failed to insert user: %v", err)
	}

	// Insert workout.
	_, err = database.Exec(
		`INSERT INTO workouts (id, user_id, name, description, intervals) VALUES (?, ?, ?, ?, ?)`,
		"w1", "u1", "Sweet Spot", "2x20", `[{"type":"steady","duration_sec":1200,"power_start":0.9,"power_end":0.9}]`,
	)
	if err != nil {
		t.Fatalf("failed to insert workout: %v", err)
	}

	var name, intervals string
	if err := database.QueryRow("SELECT name, intervals FROM workouts WHERE id = ?", "w1").Scan(&name, &intervals); err != nil {
		t.Fatalf("failed to query workout: %v", err)
	}
	if name != "Sweet Spot" {
		t.Errorf("name: got %q, want %q", name, "Sweet Spot")
	}
}
