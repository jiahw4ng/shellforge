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
		"--user", setupUser,
		"--workdir", learnerWorkspace,
		name,
		bashExecutable, "-e", "-u", "-o", "pipefail", "-c", setup,
	}
}

// getCreateContainerArguments returns the Docker flags that apply Shellforge's
// isolation limits to one newly created lesson container.
func getCreateContainerArguments(name string) []string {
	/*
		--platform linux/amd64: Force the container to run in the amd64 architecture
		--name=<name>: Give the container the unique name so concurrent lessons do not collide
		--label shellforge.managed=true: Mark the container as managed by Shellforge
		--label shellforge.session=<name>: Mark the container as belonging to this session
		--network none: Disable all networking for the container
		--memory 256m: Limit the container to 256 MiB of RAM
		--cpus 0.5: Limit the container to 50% of one CPU core
		--pids-limit 128: Limit the container to 128 processes
		image: Use the locally built lesson image
		sleep infinity: Keep the container alive until explicitly removed
	*/
	return []string{
		"create",
		"--platform", containerPlatform,
		"--name", name,
		"--label", managedContainerLabel,
		"--label", containerSessionLabelPrefix + name,
		"--network", containerNetworkMode,
		"--memory", containerMemoryLimit,
		"--cpus", containerCPULimit,
		"--pids-limit", containerPIDLimit,
		image,
		"sleep", "infinity",
	}
}
