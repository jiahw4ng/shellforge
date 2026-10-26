// Package storage persists completed lesson IDs and learner settings for the
// current user in one private SQLite database.
package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

// OpenDefault opens the database in Shellforge's per-user state directory.
func OpenDefault() (*Store, error) {
	homeDirectory, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("find user home directory: %w", err)
	}

	path := filepath.Join(homeDirectory, localStateDirectory, stateDirectory, applicationStateDirectory, databaseFileName)
	return Open(path)
}

// Open opens or creates the database at path.
func Open(path string) (*Store, error) {
	directory := filepath.Dir(path)
	if err := os.MkdirAll(directory, 0o700); err != nil {
		return nil, fmt.Errorf("create storage directory: %w", err)
	}

	database, err := sql.Open(databaseDriverName, path)
	if err != nil {
		return nil, fmt.Errorf("open storage database: %w", err)
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
		return fmt.Errorf("connect to storage database: %w", err)
	}
	if err := os.Chmod(path, 0o600); err != nil {
		return fmt.Errorf("set storage database permissions: %w", err)
	}
	if _, err := s.db.Exec(createSchema); err != nil {
		return fmt.Errorf("create storage database schema: %w", err)
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

// CursorStyle returns the saved cursor style ID and whether one has been saved.
func (s *Store) CursorStyle(ctx context.Context) (string, bool, error) {
	var styleID string
	err := s.db.QueryRowContext(ctx, `SELECT value FROM settings WHERE key = ?`, cursorStyleSettingKey).Scan(&styleID)
	if errors.Is(err, sql.ErrNoRows) {
		return "", false, nil
	}
	if err != nil {
		return "", false, fmt.Errorf("read cursor style: %w", err)
	}
	return styleID, true, nil
}

// SaveCursorStyle records the cursor style ID, replacing any earlier choice.
func (s *Store) SaveCursorStyle(ctx context.Context, styleID string) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO settings (key, value)
		VALUES (?, ?)
		ON CONFLICT (key) DO UPDATE SET value = excluded.value`, cursorStyleSettingKey, styleID)
	if err != nil {
		return fmt.Errorf("save cursor style %q: %w", styleID, err)
	}
	return nil
}

// Close closes the underlying SQLite database.
func (s *Store) Close() error {
	return s.db.Close()
}
