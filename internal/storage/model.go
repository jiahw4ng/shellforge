package storage

import (
	"context"
	"database/sql"
)

const (
	cursorStyleSettingKey     = "cursor_style"
	databaseFileName          = "completions.db"
	databaseDriverName        = "sqlite"
	localStateDirectory       = ".local"
	stateDirectory            = "state"
	applicationStateDirectory = "shellforge"
)

const (
	createSchema = `
		CREATE TABLE IF NOT EXISTS lesson_completions (
			lesson_id TEXT PRIMARY KEY NOT NULL,
			completed_at INTEGER NOT NULL
		);
		CREATE TABLE IF NOT EXISTS settings (
			key TEXT PRIMARY KEY NOT NULL,
			value TEXT NOT NULL
		);`
	// tables created:
	// lesson_completions(lesson_id PK, completed_at)
	// settings(key PK, value)
	selectCompletedLessonsQuery = `SELECT lesson_id FROM lesson_completions ORDER BY lesson_id`
	// ignore repeated completions so the original completed_at is kept
	insertCompletedLessonQuery = `
		INSERT INTO lesson_completions (lesson_id, completed_at)
		VALUES (?, unixepoch())
		ON CONFLICT (lesson_id) DO NOTHING`
	deleteCompletedLessonsQuery = `DELETE FROM lesson_completions`
	selectSettingQuery          = `SELECT value FROM settings WHERE key = ?`
	// replace any earlier value for the same key
	upsertSettingQuery = `
		INSERT INTO settings (key, value)
		VALUES (?, ?)
		ON CONFLICT (key) DO UPDATE SET value = excluded.value`
)

// Store persists lesson completion and learner settings in a local SQLite database.
type Store struct {
	db *sql.DB
}

// CompletionStore is the persistence boundary for completed lesson IDs.
// The production SQLite storage.Store implements it; app tests use a small fake.
type CompletionStore interface {
	CompletedLessonIDs(context.Context) ([]string, error)
	MarkCompleted(context.Context, string) error
	ResetLessonCompletions(context.Context) error
}

// SettingsStore is the persistence boundary for learner settings.
type SettingsStore interface {
	CursorStyle(context.Context) (string, bool, error)
	SaveCursorStyle(context.Context, string) error
}

// Storage is everything Shellforge persists. *Store implements it with one
// SQLite database; consumers keep only the narrower interface they use.
type Storage interface {
	CompletionStore
	SettingsStore
}
