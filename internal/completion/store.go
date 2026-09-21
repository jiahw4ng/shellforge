// Package completion persists completed lesson IDs for the current user.
package completion

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

// OpenDefault opens the completion database in Shellforge's per-user state directory.
func OpenDefault() (*Store, error) {
	homeDirectory, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("find user home directory: %w", err)
	}

	path := filepath.Join(homeDirectory, ".local", "state", "shellforge", databaseFileName)
	return Open(path)
}

// Open opens or creates a completion database at path.
func Open(path string) (*Store, error) {
	directory := filepath.Dir(path)
	if err := os.MkdirAll(directory, 0o700); err != nil {
		return nil, fmt.Errorf("create completion database directory: %w", err)
	}

	database, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("open completion database: %w", err)
	}
	database.SetMaxOpenConns(1)

	store := &Store{db: database}
	if err := store.initialize(path); err != nil {
		_ = database.Close()
		return nil, err
	}
	return store, nil
}

func (s *Store) initialize(path string) error {
	if err := s.db.Ping(); err != nil {
		return fmt.Errorf("connect to completion database: %w", err)
	}
	if err := os.Chmod(path, 0o600); err != nil {
		return fmt.Errorf("set completion database permissions: %w", err)
	}
	if _, err := s.db.Exec(createSchema); err != nil {
		return fmt.Errorf("create completion database schema: %w", err)
	}
	return nil
}

// CompletedLessonIDs returns every lesson ID that has been completed.
func (s *Store) CompletedLessonIDs(ctx context.Context) ([]string, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT lesson_id FROM lesson_completions ORDER BY lesson_id`)
	if err != nil {
		return nil, fmt.Errorf("query completed lessons: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var lessonIDs []string
	for rows.Next() {
		var lessonID string
		if err := rows.Scan(&lessonID); err != nil {
			return nil, fmt.Errorf("scan completed lesson: %w", err)
		}
		lessonIDs = append(lessonIDs, lessonID)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("read completed lessons: %w", err)
	}
	return lessonIDs, nil
}

// MarkCompleted records a lesson as complete. Repeated calls are idempotent.
func (s *Store) MarkCompleted(ctx context.Context, lessonID string) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO lesson_completions (lesson_id, completed_at)
		VALUES (?, unixepoch())
		ON CONFLICT (lesson_id) DO NOTHING`, lessonID)
	if err != nil {
		return fmt.Errorf("mark lesson %q complete: %w", lessonID, err)
	}
	return nil
}

// ResetLessonCompletions removes every persisted lesson completion.
func (s *Store) ResetLessonCompletions(ctx context.Context) error {
	if _, err := s.db.ExecContext(ctx, `DELETE FROM lesson_completions`); err != nil {
		return fmt.Errorf("reset lesson completions: %w", err)
	}
	return nil
}

// Close closes the underlying SQLite database.
func (s *Store) Close() error {
	return s.db.Close()
}
