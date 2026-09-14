package container

import (
	"strings"
	"testing"
)

func TestCreateArgumentsUseRequiredIsolation(t *testing.T) {
	arguments := createArguments("shellforge-test")
	joined := strings.Join(arguments, "\x00")

	for _, required := range []string{
		"--platform\x00linux/amd64",
		"--name\x00shellforge-test",
		"--network\x00none",
		"--memory\x00256m",
		"--cpus\x000.5",
		"--pids-limit\x00128",
		Image,
	} {
		if !strings.Contains(joined, required) {
			t.Errorf("container arguments do not contain %q", required)
		}
	}

	for _, forbidden := range []string{"--privileged", "--volume", "--mount"} {
		if strings.Contains(joined, forbidden) {
			t.Errorf("container arguments contain forbidden option %q", forbidden)
		}
	}
}

func TestShellCommandUsesStudentAndContainerWorkspace(t *testing.T) {
	command := (&LessonContainer{name: "shellforge-test"}).ShellCommand()
	joined := strings.Join(command.Args, "\x00")

	for _, required := range []string{
		"exec", "--interactive", "--tty", "--user\x00student",
		"--workdir\x00/home/student/workspace", "shellforge-test",
		"/bin/bash", "--rcfile\x00/opt/shellforge/interactive.bashrc",
	} {
		if !strings.Contains(joined, required) {
			t.Errorf("shell command does not contain %q", required)
		}
	}
}
