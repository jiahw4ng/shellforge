package app

import (
	"context"
	"log/slog"
	"shellforge/internal/storage"
	"shellforge/internal/ui"
	"time"

	"charm.land/bubbles/v2/spinner"
	"charm.land/bubbles/v2/viewport"
)

// startupThemeTimeout bounds the blocking theme read at startup. The database
// is local, so a read that takes longer is treated as unavailable.
const startupThemeTimeout = time.Second

// New creates the initial main-menu state without persistence.
func New() State {
	return NewWithStorage(nil)
}

// NewWithStorage creates the initial state with lesson and settings
// persistence. A nil storage keeps all data in memory only. It reads the saved
// theme synchronously and applies it, so the first frame is already themed.
func NewWithStorage(store storage.Storage) State {
	state := State{
		lessons: lessonState{
			completed: make(map[string]bool),
			store:     store,
			guide:     viewport.New(),
		},
		preferences: preferencesState{store: store},
		term:        terminalState{spinner: spinner.New(spinner.WithSpinner(spinner.Dot))},
		help:        newHelpModel(),
	}
	if store != nil {
		state.loadSavedTheme()
	}
	return state
}

// loadSavedTheme applies the persisted theme BEFORE the program starts, so the
// application never paints in the default theme and then repaints. Missing,
// unknown, or unreadable values keep the default. It runs before Bubble Tea
// takes over, so changing the process-wide theme here cannot race with View.
func (s *State) loadSavedTheme() {
	ctx, cancel := context.WithTimeout(context.Background(), startupThemeTimeout)
	defer cancel()
	themeID, found, err := s.preferences.store.Theme(ctx)
	if err != nil {
		slog.Error("Could not load theme", logErrorKey, err)
		return
	}
	if !found {
		return
	}
	theme, ok := ui.ThemeByID(themeID)
	if !ok {
		slog.Error("Ignoring unknown persisted theme", "theme", themeID)
		return
	}
	s.setTheme(theme)
}
