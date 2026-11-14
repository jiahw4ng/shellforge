package app

import (
	"context"
	"log/slog"
	"shellforge/internal/config"
	"shellforge/internal/lessons"
	"shellforge/internal/storage"
	"time"

	tea "charm.land/bubbletea/v2"
)

func (s *State) activateLessonSelection() {
	if len(s.lessons.available) == 0 {
		if s.lessons.err != nil {
			s.returnToMainMenu()
		}
		return
	}
	if s.nav.selection == len(s.lessons.available) {
		s.returnToMainMenu()
		return
	}

	s.openLesson(s.nav.selection)
}

// openLesson shows a lesson from its first page, with no hints revealed.
func (s *State) openLesson(index int) {
	s.lessons.activeIdx = index
	s.lessons.activePage = 0
	s.lessons.revealedHints = 0
	s.lessons.hintsExpanded = false
	s.lessons.guide.GotoTop()
	s.nav.screen = config.LessonTerminalScreen
}

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

type lessonsLoadedMsg struct {
	lessons []lessons.Lesson
	err     error
}

type completionsLoadedMsg struct {
	lessonIDs []string
	err       error
}

// handleLessonsLoaded records the load result and replaces the available
// lessons only when loading succeeded.
func (s State) handleLessonsLoaded(msg lessonsLoadedMsg) (tea.Model, tea.Cmd) {
	s.lessons.err = msg.err
	if msg.err == nil {
		s.lessons.available = msg.lessons
	}
	return s, nil
}

// handleCompletionsLoaded merges persisted lesson IDs into the completion map,
// leaving the current completion state unchanged when loading failed.
func (s State) handleCompletionsLoaded(msg completionsLoadedMsg) (tea.Model, tea.Cmd) {
	if msg.err != nil {
		slog.Error("Could not load lesson completion", logErrorKey, msg.err)
		s.lessons.progressNotLoaded = true
		return s, nil
	}
	for _, lessonID := range msg.lessonIDs {
		s.lessons.completed[lessonID] = true
	}
	return s, nil
}
