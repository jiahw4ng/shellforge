package completion

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

const createSchema = `
CREATE TABLE IF NOT EXISTS lesson_completions (
	lesson_id TEXT PRIMARY KEY NOT NULL,
	completed_at INTEGER NOT NULL
);
CREATE TABLE IF NOT EXISTS settings (
	key TEXT PRIMARY KEY NOT NULL,
	value TEXT NOT NULL
);`

// Store persists lesson completion and learner settings in a local SQLite database.
type Store struct {
	db *sql.DB
}

// CompletionStore is the persistence boundary used by the app state machine.
// the production SQLite completion.Store above implements this interface, but it is mocked for testing
// using fakeCompletionStore
type CompletionStore interface {
	CompletedLessonIDs(context.Context) ([]string, error)
	MarkCompleted(context.Context, string) error
	ResetLessonCompletions(context.Context) error
}

// SettingsStore is the persistence boundary for learner settings. The production
// completion.Store implements it alongside CompletionStore.
type SettingsStore interface {
	CursorStyle(context.Context) (string, bool, error)
	SaveCursorStyle(context.Context, string) error
}
