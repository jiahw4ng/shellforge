package completion

import (
	"context"
	"database/sql"
)

const databaseFileName = "completions.db"

const createSchema = `
CREATE TABLE IF NOT EXISTS lesson_completions (
	lesson_id TEXT PRIMARY KEY NOT NULL,
	completed_at INTEGER NOT NULL
);`

// Store persists lesson completion in a local SQLite database.
type Store struct {
	db *sql.DB
}

// CompletionStore is the persistence boundary used by the app state machine.
// the production SQLite Store above implements this interface, but it is mocked for testing
// using fakeCompletionStore
type CompletionStore interface {
	CompletedLessonIDs(context.Context) ([]string, error)
	MarkCompleted(context.Context, string) error
	ResetLessonCompletions(context.Context) error
}
