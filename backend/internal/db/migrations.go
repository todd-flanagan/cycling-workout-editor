package db

import (
	"database/sql"
	"fmt"
	"log"
)

// Migration represents a single database migration.
type Migration struct {
	Version     int
	Description string
	SQL         string
}

// migrations is the ordered list of all migrations.
var migrations = []Migration{
	{
		Version:     1,
		Description: "Create users, ftp_history, and workouts tables",
		SQL: `
CREATE TABLE IF NOT EXISTS users (
	id TEXT PRIMARY KEY,
	google_id TEXT UNIQUE NOT NULL,
	email TEXT NOT NULL,
	display_name TEXT NOT NULL,
	ftp INTEGER NOT NULL DEFAULT 200,
	created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS ftp_history (
	id TEXT PRIMARY KEY,
	user_id TEXT NOT NULL REFERENCES users(id),
	ftp INTEGER NOT NULL,
	recorded_at DATE NOT NULL,
	source TEXT,
	created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS workouts (
	id TEXT PRIMARY KEY,
	user_id TEXT NOT NULL REFERENCES users(id),
	name TEXT NOT NULL DEFAULT 'Untitled Workout',
	description TEXT DEFAULT '',
	intervals TEXT NOT NULL DEFAULT '[]',
	template_id TEXT,
	created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_ftp_history_user_id ON ftp_history(user_id);
CREATE INDEX IF NOT EXISTS idx_workouts_user_id ON workouts(user_id);
`,
	},
}

// RunMigrations executes all pending migrations in order.
func RunMigrations(db *sql.DB) error {
	// Create migrations tracking table.
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version INTEGER PRIMARY KEY,
			description TEXT NOT NULL,
			applied_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		return fmt.Errorf("create schema_migrations table: %w", err)
	}

	for _, m := range migrations {
		var count int
		err := db.QueryRow("SELECT COUNT(*) FROM schema_migrations WHERE version = ?", m.Version).Scan(&count)
		if err != nil {
			return fmt.Errorf("check migration %d: %w", m.Version, err)
		}
		if count > 0 {
			continue // Already applied.
		}

		log.Printf("Applying migration %d: %s", m.Version, m.Description)

		tx, err := db.Begin()
		if err != nil {
			return fmt.Errorf("begin tx for migration %d: %w", m.Version, err)
		}

		if _, err := tx.Exec(m.SQL); err != nil {
			tx.Rollback()
			return fmt.Errorf("exec migration %d: %w", m.Version, err)
		}

		if _, err := tx.Exec(
			"INSERT INTO schema_migrations (version, description) VALUES (?, ?)",
			m.Version, m.Description,
		); err != nil {
			tx.Rollback()
			return fmt.Errorf("record migration %d: %w", m.Version, err)
		}

		if err := tx.Commit(); err != nil {
			return fmt.Errorf("commit migration %d: %w", m.Version, err)
		}

		log.Printf("Migration %d applied successfully", m.Version)
	}

	return nil
}
