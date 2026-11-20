package app

import (
	"context"
	"log/slog"
	"shellforge/internal/config"
	"shellforge/internal/lessons"
	"shellforge/internal/screens"
	"shellforge/internal/storage"
	"time"

	tea "charm.land/bubbletea/v2"
)

// openLesson opens the lesson at index, the one thing the lesson screen
// cannot work out itself, and returns the work its onArrival starts. The
// lesson list stays open underneath, so leaving the lesson returns to it in
// the list.
func (s *State) openLesson(index int) tea.Cmd {
	s.lessons.activeIdx = index
	return s.openChildScreen(config.LessonTerminalScreen)
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

// lessonsBehaviour is the lesson list's screen behaviour: every lesson, then
// Back.
func lessonsBehaviour() screenBehaviour {
	return screenBehaviour{
		parent: new(config.MainMenuScreen),
		list: &listBehaviour{
			// Every lesson, then Back. While lessons are still loading there is
			// nothing to select.
			itemCount: func(s State) int {
				if len(s.lessons.available) == 0 && s.lessons.err == nil {
					return 0
				}
				return len(s.lessons.available) + 1
			},
			// Enter opens the selected lesson or returns to the main menu from
			// Back. If loading failed, Enter returns to the menu as well.
			onSelectItem: func(s *State) tea.Cmd {
				if len(s.lessons.available) == 0 {
					if s.lessons.err != nil {
						return s.goBackToParentScreen()
					}
					return nil
				}
				if s.nav.selection == len(s.lessons.available) {
					return s.goBackToParentScreen()
				}

				return s.openLesson(s.nav.selection)
			},
		},
		view: func(s State, helpWidth int) string {
			return screens.LessonList(s.lessons.available, s.lessons.completed, s.nav.selection, s.lessons.err, s.navigationHelp(helpWidth))
		},
		help: helpBehaviour{
			hasHelp: true,
		},
	}
}
