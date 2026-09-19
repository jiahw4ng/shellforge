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
	command := (&Container{Name: "shellforge-test"}).ShellCommand()
	joined := strings.Join(command.Args, "\x00")

	for _, required := range []string{
		"exec", "--interactive", "--tty", "--user\x00student",
		"--workdir\x00/home/student/workspace", "shellforge-test",
		"/bin/bash", "--noprofile", "--rcfile\x00/etc/shellforge/bashrc", "-i",
	} {
		if !strings.Contains(joined, required) {
			t.Errorf("shell command does not contain %q", required)
		}
	}
}

// TestSetupCommandUsesRootAndStrictNonInteractiveBash verifies setup can
// prepare student-owned files without allocating a second terminal.
func TestSetupCommandUsesRootAndStrictNonInteractiveBash(t *testing.T) {
	arguments := setupCommandArguments("shellforge-test", "mkdir -p project")
	joined := strings.Join(arguments, "\x00")

	for _, required := range []string{
		"exec", "--user\x00root", "--workdir\x00/home/student/workspace",
		"shellforge-test", "/bin/bash", "-e", "-u", "-o", "pipefail", "-c", "mkdir -p project",
	} {
		if !strings.Contains(joined, required) {
			t.Errorf("setup command does not contain %q", required)
		}
	}

	for _, forbidden := range []string{"--interactive", "--tty"} {
		if strings.Contains(joined, forbidden) {
			t.Errorf("setup command contains forbidden option %q", forbidden)
		}
	}
}
