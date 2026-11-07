package app

import (
	"errors"
	"shellforge/internal/config"
	"shellforge/internal/lessons"
	"testing"
)

func TestBeginAttemptStartsAFreshGenerationAndClearsFailure(t *testing.T) {
	term := terminalState{gen: 4, output: "old", err: errors.New("old"), hasRequestedExit: true}

	term.beginAttempt()

	if term.gen != 5 || !term.isStarting {
		t.Errorf("gen = %d, isStarting = %v; want a new generation that is starting", term.gen, term.isStarting)
	}
	if term.output != "" || term.err != nil || term.hasRequestedExit {
		t.Errorf("attempt kept old state: output=%q err=%v requestedExit=%v", term.output, term.err, term.hasRequestedExit)
	}
	if len(term.spinner.Spinner.Frames) == 0 {
		t.Error("attempt did not create a spinner")
	}
}

func TestClearFailureForgetsOnlyOutputAndError(t *testing.T) {
	term := terminalState{gen: 2, output: "frame", err: errors.New("boom"), isStarting: true}

	term.clearFailure()

	if term.output != "" || term.err != nil {
		t.Fatalf("clearFailure left output=%q err=%v", term.output, term.err)
	}
	if term.gen != 2 || !term.isStarting {
		t.Fatalf("clearFailure changed unrelated state: gen=%d isStarting=%v", term.gen, term.isStarting)
	}
}

func TestCompleteLessonPersistsOnlyTheFirstCompletion(t *testing.T) {
	store := &fakeStorage{}
	model := NewWithStorage(store)

	if command := model.completeLesson("lesson-id"); command == nil {
		t.Fatal("first completion returned no save command")
	}
	if !model.lessons.completed["lesson-id"] {
		t.Fatal("first completion was not recorded in memory")
	}
	if command := model.completeLesson("lesson-id"); command != nil {
		t.Fatal("repeat completion scheduled another save")
	}
}

func TestCompleteLessonWithoutStorageStillMarksTheLessonComplete(t *testing.T) {
	model := New()

	if command := model.completeLesson("lesson-id"); command != nil {
		t.Fatal("completion without storage returned a save command")
	}
	if !model.lessons.completed["lesson-id"] {
		t.Fatal("completion without storage was not recorded in memory")
	}
}

func TestOpenLessonStartsFromTheFirstPageWithNothingRevealed(t *testing.T) {
	model := State{lessons: lessonState{
		available:     []lessons.Lesson{{}, {}},
		activeIdx:     0,
		activePage:    3,
		revealedHints: 2,
		hintsExpanded: true,
	}}

	model.openLesson(1)

	if model.lessons.activeIdx != 1 || model.lessons.activePage != 0 || model.lessons.revealedHints != 0 || model.lessons.hintsExpanded {
		t.Fatalf("lesson state after openLesson = %#v", model.lessons)
	}
	if model.nav.screen != config.LessonTerminalScreen {
		t.Fatalf("screen = %d, want the lesson terminal screen", model.nav.screen)
	}
}
