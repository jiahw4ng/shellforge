package container

// getExecCommandArguments builds a non-interactive docker exec invocation.
func getExecCommandArguments(name string, user string, workingDir string, command []string) []string {
	arguments := []string{
		"exec",
		"--user", user,
		"--workdir", workingDir,
		name,
	}
	return append(arguments, command...)
}

// getSetupCommandArguments returns a non-interactive root command because lesson
// setup creates the initial filesystem before the learner Bash starts.
func getSetupCommandArguments(name, setup string) []string {
	return []string{
		"exec",
		"--user", "root",
		"--workdir", "/home/student/workspace",
		name,
		"/bin/bash", "-e", "-u", "-o", "pipefail", "-c", setup,
	}
}

// getCreateContainerArguments returns the Docker flags that apply Shellforge's
// isolation limits to one newly created lesson container.
func getCreateContainerArguments(name string) []string {
	return []string{
		"create",
		"--platform", "linux/amd64",
		"--name", name,
		"--label", "shellforge.managed=true",
		"--label", "shellforge.session=" + name,
		"--network", "none",
		"--memory", "256m",
		"--cpus", "0.5",
		"--pids-limit", "128",
		image,
		"sleep", "infinity",
	}
}
