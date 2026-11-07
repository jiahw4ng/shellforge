package app

import (
	"errors"
	"shellforge/internal/config"
	"shellforge/internal/container"
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
)

// terminalRoutes renders the same terminal state through the standalone sandbox
// and through the lesson pane, which decide their content separately.
func terminalRoutes(t *testing.T, configure func(*State)) map[string]string {
	t.Helper()
	sandbox := New()
	sandbox.nav.screen = config.SandboxTerminalScreen
	lesson := loadedState(t)
	lesson.nav.screen = config.LessonTerminalScreen

	views := map[string]string{}
	for name, model := range map[string]State{"sandbox": sandbox, "lesson": lesson} {
		model.viewport = viewportState{width: 220, height: 40}
		configure(&model)
		views[name] = ansi.Strip(model.View().Content)
	}
	return views
}

func TestTerminalPaneShowsTheSameContentInSandboxAndLesson(t *testing.T) {
	tests := []struct {
		name      string
		configure func(*State)
		want      string
		wantError bool
	}{
		{"starting", func(s *State) { s.term.isStarting = true }, "Starting sandboxed shell", false},
		{"recognised startup problem", func(s *State) { s.term.err = dockerStartFailure(container.ReasonDockerUnreachable) }, "Docker isn't reachable.", true},
		{"unrecognised startup error", func(s *State) { s.term.err = errors.New("pty failed") }, "pty failed", true},
		{"exit with output", func(s *State) {
			s.term.err = dockerStartFailure(container.ReasonDockerUnreachable)
			s.term.output = "bash: exited unexpectedly"
		}, "bash: exited unexpectedly", true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			for route, view := range terminalRoutes(t, test.configure) {
				if !strings.Contains(view, test.want) {
					t.Errorf("%s view does not contain %q:\n%s", route, test.want, view)
				}
				if got := strings.Contains(view, "Oops! There was an error."); got != test.wantError {
					t.Errorf("%s view shows the error title = %v, want %v:\n%s", route, got, test.wantError, view)
				}
			}
		})
	}
}

func TestExitOutputHidesTheStartupProblemInBothRoutes(t *testing.T) {
	views := terminalRoutes(t, func(s *State) {
		s.term.err = dockerStartFailure(container.ReasonDockerUnreachable)
		s.term.output = "bash: exited unexpectedly"
	})
	for route, view := range views {
		if strings.Contains(view, "Docker isn't reachable.") {
			t.Errorf("%s view shows the startup problem next to the terminal's own output:\n%s", route, view)
		}
	}
}
