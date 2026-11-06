package app

import (
	"errors"
	"shellforge/internal/config"
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
)

// healthyStorage returns a state whose persistence works, so no notice shows.
func healthyStorage(model State) State {
	model.lessons.store = &fakeStorage{}
	return model
}

// lineAbove returns the stripped view line directly above the one holding marker.
func lineAbove(t *testing.T, view, marker string) string {
	t.Helper()
	lines := strings.Split(view, "\n")
	for index, line := range lines {
		if strings.Contains(line, marker) {
			if index == 0 {
				t.Fatalf("%q is on the first line, so nothing is above it:\n%s", marker, view)
			}
			return lines[index-1]
		}
	}
	t.Fatalf("view does not contain %q:\n%s", marker, view)
	return ""
}

func TestStorageNoticeStatesAndPriority(t *testing.T) {
	healthy := healthyStorage(New())
	if got := healthy.storageNotice(); got != "" {
		t.Fatalf("storageNotice() with working storage = %q, want none", got)
	}

	loadFailed := healthy
	loadFailed.lessons.progressNotLoaded = true
	if got := loadFailed.storageNotice(); got != progressNotLoadedNotice {
		t.Errorf("load failure notice = %q, want %q", got, progressNotLoadedNotice)
	}

	bothFailed := loadFailed
	bothFailed.lessons.progressNotSaved = true
	if got := bothFailed.storageNotice(); got != progressNotSavedNotice {
		t.Errorf("save failure must outrank load failure: got %q, want %q", got, progressNotSavedNotice)
	}

	unavailable := bothFailed
	unavailable.lessons.store = nil
	if got := unavailable.storageNotice(); got != storageUnavailableNotice {
		t.Errorf("missing storage must outrank everything: got %q, want %q", got, storageUnavailableNotice)
	}
}

func TestStorageNoticeSitsDirectlyAboveHelpOnEveryHelpScreen(t *testing.T) {
	for name, screen := range map[string]config.Screen{
		"main menu":          config.MenuScreen,
		"lesson list":        config.LessonsScreen,
		"settings":           config.SettingsScreen,
		"cursor style":       config.CursorStyleScreen,
		"theme":              config.ThemeScreen,
		"reset confirmation": config.ResetConfirmationScreen,
		"lesson":             config.LessonTerminalScreen,
	} {
		t.Run(name, func(t *testing.T) {
			model := loadedState(t) // no store: persistence is unavailable
			model.nav.screen = screen
			model.viewport = viewportState{width: 140, height: 40}

			view := ansi.Strip(model.View().Content)
			if above := lineAbove(t, view, "F3 Help"); !strings.Contains(above, "when you quit.") {
				t.Fatalf("line above the help line = %q, want the end of the storage notice:\n%s", above, view)
			}

			healthy := healthyStorage(model)
			if view := ansi.Strip(healthy.View().Content); strings.Contains(view, "can't be saved") {
				t.Fatalf("working storage still shows the notice:\n%s", view)
			}
		})
	}
}

func TestStorageNoticeHidesWhileResetting(t *testing.T) {
	model := loadedState(t)
	model.nav.screen = config.ResetConfirmationScreen
	model.settings.isResetting = true
	model.viewport = viewportState{width: 140, height: 40}

	if view := ansi.Strip(model.View().Content); strings.Contains(view, "can't be saved") || strings.Contains(view, "F3 Help") {
		t.Fatalf("reset screen shows help or the notice while resetting:\n%s", view)
	}
}

func TestFailedCompletionSaveShowsNoticeUntilQuit(t *testing.T) {
	model := healthyStorage(New())
	model.viewport = viewportState{width: 140, height: 40}

	model = updateModel(t, model, completionSavedMsg{lessonID: "lesson-id"})
	if got := model.storageNotice(); got != "" {
		t.Fatalf("a successful save produced the notice %q", got)
	}

	model = updateModel(t, model, completionSavedMsg{lessonID: "lesson-id", err: errors.New("database unavailable")})
	view := ansi.Strip(model.View().Content)
	if above := lineAbove(t, view, "F3 Help"); !strings.Contains(above, "when you quit.") {
		t.Fatalf("line above the help line = %q, want the not-saved notice:\n%s", above, view)
	}
	if !strings.Contains(view, "Your lesson progress couldn't be saved.") {
		t.Fatalf("view does not show the not-saved notice:\n%s", view)
	}

	// a later success must not hide it: the earlier completion is still unsaved
	model = updateModel(t, model, completionSavedMsg{lessonID: "other-lesson"})
	if got := model.storageNotice(); got != progressNotSavedNotice {
		t.Fatalf("notice after a later successful save = %q, want it kept", got)
	}
}

func TestFailedCompletionLoadShowsNotice(t *testing.T) {
	model := healthyStorage(New())
	model.viewport = viewportState{width: 140, height: 40}

	model = updateModel(t, model, completionsLoadedMsg{err: errors.New("database unavailable")})
	view := ansi.Strip(model.View().Content)
	if above := lineAbove(t, view, "F3 Help"); !strings.Contains(above, "show a checkmark.") {
		t.Fatalf("line above the help line = %q, want the not-loaded notice:\n%s", above, view)
	}
}

func TestSuccessfulResetClearsNoticeButFailedResetKeepsIt(t *testing.T) {
	model := healthyStorage(New())
	model.lessons.progressNotSaved = true
	model.lessons.progressNotLoaded = true

	failed := updateModel(t, model, completionsResetMsg{err: errors.New("database unavailable")})
	if got := failed.storageNotice(); got != progressNotSavedNotice {
		t.Fatalf("notice after a failed reset = %q, want it kept", got)
	}

	reset := updateModel(t, model, completionsResetMsg{})
	if got := reset.storageNotice(); got != "" {
		t.Fatalf("notice after a successful reset = %q, want none", got)
	}
}

func TestStorageNoticeWrapsInsideNarrowLessonPane(t *testing.T) {
	model := loadedState(t)
	model.nav.screen = config.LessonTerminalScreen
	model.viewport = viewportState{width: 80, height: 30}

	view := ansi.Strip(model.View().Content)
	for index, line := range strings.Split(view, "\n") {
		if width := ansi.StringWidth(line); width > model.viewport.width {
			t.Fatalf("line %d is %d columns wide in an %d-column viewport:\n%s", index, width, model.viewport.width, view)
		}
	}
	for _, word := range []string{"Progress", "settings", "saved.", "quit."} {
		if !strings.Contains(view, word) {
			t.Errorf("narrow lesson view lost part of the notice (%q):\n%s", word, view)
		}
	}
	if !strings.Contains(view, "F3 Help") {
		t.Fatalf("narrow lesson view lost the help line:\n%s", view)
	}
}
