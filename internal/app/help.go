package app

import (
	"strings"

	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
)

// navigationHelp renders the shortcuts shared by non-terminal menu screens.
func (s State) navigationHelp() string {
	return renderSingleHelpLine([]key.Binding{
		keys.up,
		keys.down,
		keys.selectItem,
		keys.quit,
	})
}

// lessonHelp renders lesson shortcuts in small related groups so the help
// remains readable inside the left half of the lesson screen.
func (s State) lessonHelp() string {
	// only enable the previous page help prompt
	// if the user is not on the first page of the lesson
	previousPage := keys.previousPage
	previousPage.SetEnabled(s.lessons.activePage > 0)

	// only enable the next page help prompt
	// if the user is not on the last page of the lesson
	nextPage := keys.nextPage
	nextPage.SetEnabled(
		s.lessons.activeIdx >= 0 &&
			s.lessons.activeIdx < len(s.lessons.available) &&
			s.lessons.activePage < len(s.lessons.available[s.lessons.activeIdx].Pages)-1,
	)

	// only enable the reset sandbox help prompt
	// if the user is not in the process of starting a new terminal session
	resetSandbox := keys.resetSandbox
	resetSandbox.SetEnabled(
		!s.term.isStarting &&
			s.lessons.activeIdx >= 0 &&
			s.lessons.activeIdx < len(s.lessons.available),
	)

	// only enable the check progress help prompt
	// if the user is in a terminal session and not already checking progress
	checkProgress := keys.checkProgress
	checkProgress.SetEnabled(s.term.session != nil && !s.lessons.progress.isChecking)

	// only enable the return back help prompt
	// if the user is in a terminal session or has an error to return from
	returnBack := keys.returnBack
	returnBack.SetEnabled(s.term.session != nil || s.term.err != nil)

	// group the lesson help prompts into related groups for better readability
	keyBindingGroups := [][]key.Binding{
		{previousPage, nextPage},
		{keys.guidePageUp, keys.guidePageDown},
		{resetSandbox},
		{checkProgress},
		{returnBack},
	}

	// render each group of help prompts into a single line and join the lines together
	helpLines := make([]string, len(keyBindingGroups))
	for i, keyBindingGroup := range keyBindingGroups {
		if line := renderSingleHelpLine(keyBindingGroup); line != "" {
			helpLines[i] = line
		}
	}
	return strings.Join(helpLines, "\n")
}

func renderSingleHelpLine(bindings []key.Binding) string {
	return help.New().ShortHelpView(bindings)
}
