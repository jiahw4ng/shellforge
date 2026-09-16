package container

import (
	"strings"
	"testing"
)

// TestCreateArgumentsUseRequiredIsolation ensures every lesson gets the
// intended Docker limits and cannot receive host filesystem mounts.
func TestCreateArgumentsUseRequiredIsolation(t *testing.T) {
	arguments := getCreateContainerArguments("shellforge-test")
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

// TestShellCommandUsesStudentAndContainerWorkspace checks the attached Bash
// session starts as the learner in the container-only workspace.
func TestShellCommandUsesStudentAndContainerWorkspace(t *testing.T) {
	command := (&LessonContainer{Name: "shellforge-test"}).ShellCommand()
	joined := strings.Join(command.Args, "\x00")

	for _, required := range []string{
		"exec", "--interactive", "--tty", "--user\x00student",
		"--workdir\x00/home/student/workspace", "shellforge-test",
		"/bin/bash", "--noprofile", "--norc", "-i",
	} {
		if !strings.Contains(joined, required) {
			t.Errorf("shell command does not contain %q", required)
		}
	}
}
