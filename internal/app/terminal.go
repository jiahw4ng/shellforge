package app

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"shellforge/internal/config"
	"shellforge/internal/container"
	"shellforge/internal/lessons"
	"shellforge/internal/logging"
	"shellforge/internal/screens"
	"shellforge/internal/terminal"
	"shellforge/internal/ui"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
)

// handleTerminalKeyPress forwards input to an active session, or clears a
// terminal failure while returning to its parent screen. Every key reaches a
// running shell, Ctrl+D included: only the shell exiting leaves the terminal.
// note: this is used by both the lesson and standalone sandbox terminal.
func (s State) handleTerminalKeyPress(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	if s.term.session != nil {
		return s, s.term.session.Update(msg)
	}
	if key.Matches(msg, config.Keys.Quit) {
		return s, tea.Quit
	}
	if key.Matches(msg, config.Keys.SelectItem) && s.term.err != nil {
		back := s.returnFromTerminal()
		s.term.clearFailure()
		return s, back
	}
	return s, nil
}

// startTerminal clears session-only state and starts a new terminal attempt.
// Every attempt receives a generation so delayed events from an older session
// cannot alter this one.
func (s *State) startTerminal(lesson *lessons.Lesson) tea.Cmd {
	s.term.beginAttempt()
	s.lessons.progress = progressState{}
	width, height := s.terminalDimensions()
	s.term.startEvents = launchTerminal(width, height, lesson, s.term.gen)
	return tea.Batch(waitForStartEvent(s.term.startEvents), s.term.spinner.Tick)
}

// returnFromTerminal returns to the screen that launched the terminal, with
// the item that launched it still selected.
func (s *State) returnFromTerminal() tea.Cmd {
	if s.nav.screen == config.LessonTerminalScreen {
		return s.navigate(config.LessonsScreen, s.nav.selection)
	}
	return s.navigate(config.MenuScreen, s.nav.selection)
}

func (s State) usesTerminal() bool {
	return s.nav.screen == config.SandboxTerminalScreen || s.nav.screen == config.LessonTerminalScreen
}

func (s State) resizeTerminal() tea.Cmd {
	width, height := s.terminalDimensions()
	return s.term.session.Resize(width, height)
}

func (s State) terminalDimensions() (int, int) {
	width, height := ui.ApplicationContentDimensions(s.viewport.width, s.viewport.height)
	if s.nav.screen == config.LessonTerminalScreen {
		return screens.LessonTerminalDimensions(width, height)
	}
	return ui.DimensionWithFallback(width, 80), screens.TerminalContentHeight(height)
}

func (s *State) CloseTerminal() {
	if s.term.session != nil {
		s.term.session.Close()
	}
	s.term.session = nil
}

func (s State) sandboxView() string {
	width, height := ui.ApplicationContentDimensions(s.viewport.width, s.viewport.height)
	return screens.SandboxTerminalPane(s.sandboxTerminalContent(), width, height)
}

// sandboxTerminalContent is what fills the standalone sandbox: the live
// terminal, the startup spinner, or the reason the terminal is not running.
func (s State) sandboxTerminalContent() string {
	if s.term.session != nil {
		return s.term.session.View()
	}
	if s.term.isStarting {
		return screens.TerminalStart(nil, s.term.startSteps, s.term.spinner.View())
	}
	if content, ok := s.terminalStoppedView(); ok {
		return content
	}
	return screens.TerminalFailure(s.term.output, s.term.err)
}

// terminalStoppedView explains why a terminal that is not running stopped:
// whatever it printed before exiting, otherwise a recognised startup problem.
// ok is false when there is nothing specific to show.
func (s State) terminalStoppedView() (string, bool) {
	if s.term.output != "" {
		return screens.TerminalFailure(s.term.output, s.term.err), true
	}
	return s.terminalProblemView()
}

// terminalProblemView explains a recognised startup failure with a summary,
// a remedy, and how to leave. ok is false when there is no start failure or it
// is one with no specific advice, so the caller shows the error text as is.
func (s State) terminalProblemView() (string, bool) {
	if s.term.err == nil || s.term.output != "" {
		return "", false
	}
	message, ok := terminal.Describe(s.term.err)
	if !ok {
		return "", false
	}
	return screens.TerminalProblem{
		Summary: message.Summary,
		Remedy:  message.Remedy,
		Hint:    fmt.Sprintf(terminalProblemHintFormat, config.Keys.SelectItem.Help().Key),
		Detail:  message.Detail,
		LogPath: logging.Path(),
	}.View(), true
}

// handleTerminalStartProgress records the next startup step of the current
// attempt and waits for the one after it.
func (s State) handleTerminalStartProgress(msg termStartProgressMsg) (tea.Model, tea.Cmd) {
	// Steps from an older attempt are ignored.
	if msg.gen != s.term.gen || !s.term.isStarting {
		return s, nil
	}
	// append the new step to the list of steps so far, and wait for the next one
	s.term.startSteps = append(s.term.startSteps, msg.step)
	return s, waitForStartEvent(s.term.startEvents)
}

// handleTerminalStarted ignores stale results, records startup failures, or
// installs the new session and resets terminal-scoped lesson progress.
func (s State) handleTerminalStarted(msg termStartedMsg) (tea.Model, tea.Cmd) {
	// Steps from an older attempt are ignored.
	if msg.gen != s.term.gen {
		return s, nil
	}
	s.term.isStarting = false
	// clear the starting steps
	s.term.startSteps = nil
	if msg.err != nil {
		s.term.err = msg.err
		return s, nil
	}
	if msg.session == nil {
		s.term.err = terminal.NewInvalidStartResultError()
		return s, nil
	}
	s.term.session = msg.session
	s.term.output = ""
	s.lessons.progress = progressState{}
	return s, tea.Batch(s.term.session.Init(), waitForTerminalExit(s.term.session, s.term.gen), s.resizeTerminal())
}

// handleTerminalExited ignores stale exits, stops in-flight terminal work, and
// either returns from a shell the learner left (its sandbox is still running)
// or records an unexpected failure.
func (s State) handleTerminalExited(msg termExitedMsg) (tea.Model, tea.Cmd) {
	if msg.gen != s.term.gen {
		return s, nil
	}
	s.term.isStarting = false
	s.lessons.progress.isChecking = false
	if msg.sandboxRunning {
		back := s.finishShellExit()
		return s, back
	}
	s.recordUnexpectedExit()
	return s, nil
}

// finishShellExit closes the terminal and returns to the screen that
// launched it, leaving no failure behind.
func (s *State) finishShellExit() tea.Cmd {
	s.CloseTerminal()
	s.term.clearFailure()
	return s.returnFromTerminal()
}

// recordUnexpectedExit keeps the terminal's last frame for the learner, closes
// it, and records the failure. The frame is read first: closing drops the session.
func (s *State) recordUnexpectedExit() {
	s.term.output = s.term.session.View()
	s.CloseTerminal()
	s.term.err = errors.New("the terminal closed unexpectedly")
}

// terminalProblemHintFormat tells the learner how to leave a startup failure.
const terminalProblemHintFormat = "Then press %s to go back and try again."

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

// waitForTerminalExit converts the session's exit signal into a Bubble Tea message.
// <- exited: when the shell exits and channel closes, this unblocks
// and returns the termExitedMsg to the Bubble Tea event loop.
// It then asks Docker whether the sandbox is still running, before anything
// removes it, so the handler can tell a learner who typed exit from a failure.
// An unknown state counts as not running, which shows the failure screen.
func waitForTerminalExit(session *terminal.TermSession, generation uint64) tea.Cmd {
	return func() tea.Msg {
		<-session.Exited
		running, err := session.SandboxIsRunning(context.Background())
		if err != nil {
			slog.Warn("Could not check the sandbox after its shell exited", logErrorKey, err)
		}
		return termExitedMsg{gen: generation, sandboxRunning: running}
	}
}

type termStartedMsg struct {
	session *terminal.TermSession
	err     error
	gen     uint64
}

// termStartProgressMsg reports that a start attempt has reached a new step.
type termStartProgressMsg struct {
	step container.Step
	gen  uint64
}

// termExitedMsg reports that a session's shell ended. sandboxRunning is true
// only when Docker confirmed the container was still running afterwards,
// meaning the learner left the shell rather than the sandbox failing.
type termExitedMsg struct {
	gen            uint64
	sandboxRunning bool
}
