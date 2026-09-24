package app

import (
	"context"
	"shellforge/internal/assertion"
	"shellforge/internal/completion"
	"shellforge/internal/lessons"
	"shellforge/internal/terminal"
	"time"

	tea "charm.land/bubbletea/v2"
)

func loadLessons() tea.Cmd {
	return func() tea.Msg {
		loaded, err := lessons.Load()
		return LessonsLoadedMsg{Lessons: loaded, Err: err}
	}
}

func loadCompletedLessons(store completion.CompletionStore) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		lessonIDs, err := store.CompletedLessonIDs(ctx)
		return CompletionsLoadedMsg{LessonIDs: lessonIDs, Err: err}
	}
}

// startTerminal creates a terminal session without blocking Bubble Tea's event loop.
// it will load the lesson, if any
func startTerminal(width, height int, lesson *lessons.Lesson, generation uint64) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()
		session, err := terminal.Start(ctx, width, height, lesson)
		return TerminalStartedMsg{Session: session, Err: err, Generation: generation}
	}
}

// checkAssertions evaluates lesson outcomes without blocking the TUI or PTY.
func checkAssertions(session *terminal.TermSession, lessonID string, assertions []assertion.Assertion, generation uint64) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return AssertionsCheckedMsg{LessonID: lessonID, Results: session.EvaluateAssertions(ctx, assertions), Generation: generation}
	}
}

func saveCompletion(store completion.CompletionStore, lessonID string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		err := store.MarkCompleted(ctx, lessonID)
		return CompletionSavedMsg{LessonID: lessonID, Err: err}
	}
}

func resetLessonCompletions(store completion.CompletionStore) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		err := store.ResetLessonCompletions(ctx)
		return CompletionsResetMsg{Err: err}
	}
}

// waitForTerminalExit converts the session's exit signal into a Bubble Tea message.
func waitForTerminalExit(exited <-chan struct{}, generation uint64) tea.Cmd {
	return func() tea.Msg {
		<-exited
		return TerminalExitedMsg{Generation: generation}
	}
}
