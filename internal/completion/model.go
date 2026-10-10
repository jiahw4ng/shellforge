package completion

import (
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
