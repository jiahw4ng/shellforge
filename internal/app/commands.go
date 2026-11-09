package app

import (
	"context"
	"shellforge/internal/assertion"
	"shellforge/internal/container"
	"shellforge/internal/lessons"
	"shellforge/internal/storage"
	"shellforge/internal/terminal"
	"time"

	tea "charm.land/bubbletea/v2"
)

func loadLessons() tea.Cmd {
	return func() tea.Msg {
		loaded, err := lessons.Load()
		return lessonsLoadedMsg{lessons: loaded, err: err}
	}
}

func loadCompletedLessons(store storage.CompletionStore) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		lessonIDs, err := store.GetCompletedLessonIDs(ctx)
		return completionsLoadedMsg{lessonIDs: lessonIDs, err: err}
	}
}

func saveCursorStyle(store storage.SettingsStore, styleID string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		err := store.UpdateCursorStyle(ctx, styleID)
		return cursorStyleSavedMsg{styleID: styleID, err: err}
	}
}

func saveTheme(store storage.SettingsStore, themeID string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		err := store.UpdateSaveTheme(ctx, themeID)
		return themeSavedMsg{themeID: themeID, err: err}
	}
}

// startEventBuffer holds every message one start attempt can send: one per
// startup step plus the final result. Sends therefore never block, even when
// the app has moved on and stopped reading.
const startEventBuffer = 16

// launchTerminal starts a terminal session in the background without blocking
// Bubble Tea's event loop, loading the lesson's setup if there is one. The
// returned channel carries a termStartProgressMsg for each startup step,
// then one termStartedMsg, and is then closed.
func launchTerminal(width, height int, lesson *lessons.Lesson, generation uint64) <-chan tea.Msg {
	events := make(chan tea.Msg, startEventBuffer)
	// start running this in the meantime
	go func() {
		// when goroutine exits, close channel
		defer close(events)
		// the "phone number" that when called puts a message into the channel
		// THIS IS THE ProgressFunc!!
		// it closes over the events channel and the generation number,
		// so it can send messages to the right place.
		report := func(step container.Step) {
			events <- termStartProgressMsg{step: step, gen: generation}
		}
		// Each Docker step has its own timeout, so no overall deadline is set.
		session, err := terminal.Start(context.Background(), width, height, lesson, report)
		// LAST STEP: terminal has started
		events <- termStartedMsg{session: session, err: err, gen: generation}
	}()
	return events
}

// waitForStartEvent converts the next event of a start attempt into a Bubble
// Tea message. The handler for a progress message calls it again for the next one.
// basically this blocks on <-events and returns the next message all the way to state.Update()
// which then routes it to State.handleTerminalStartProgress() / State.handleTerminalStarted()
// but this wont block the UI because Bubble Tea's runtime runs
// the returned function in its own goroutine. That goroutine
// is the one that waits on <- events.
func waitForStartEvent(events <-chan tea.Msg) tea.Cmd {
	return func() tea.Msg {
		return <-events
	}
}

// checkAssertions evaluates lesson outcomes without blocking the TUI or PTY.
func checkAssertions(session *terminal.TermSession, lessonID string, assertions []assertion.Assertion, generation uint64) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return assertionsCheckedMsg{lessonID: lessonID, results: session.EvaluateAssertions(ctx, assertions), gen: generation}
	}
}

func saveCompletion(store storage.CompletionStore, lessonID string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		err := store.MarkLessonIDAsCompleted(ctx, lessonID)
		return completionSavedMsg{lessonID: lessonID, err: err}
	}
}

func resetLessonCompletions(store storage.CompletionStore) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		err := store.ResetLessonCompletions(ctx)
		return completionsResetMsg{err: err}
	}
}

// waitForTerminalExit converts the session's exit signal into a Bubble Tea message.
// <- exited: when the shell exits and channel closes, this unblocks
// and returns the termExitedMsg to the Bubble Tea event loop.
func waitForTerminalExit(exited <-chan struct{}, generation uint64) tea.Cmd {
	return func() tea.Msg {
		<-exited
		return termExitedMsg{gen: generation}
	}
}
