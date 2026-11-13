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

// getListManagedArguments lists every container, running or stopped, that
// carries Shellforge's managed label, one tab-separated line per container.
func getListManagedArguments() []string {
	return []string{
		"ps", "--all",
		"--filter", "label=" + managedContainerLabel,
		"--format", managedListFormat,
	}
}

// getCreateContainerArguments returns the Docker flags that apply Shellforge's
// isolation limits to one newly created lesson container.
func getCreateContainerArguments(name string) []string {
	/*
		--name=<name>: Give the container the unique name so concurrent lessons do not collide
		--label shellforge.managed=true: Mark the container as managed by Shellforge
		--label shellforge.session=<name>: Mark the container as belonging to this session
		--network none: Disable all networking for the container
		--memory 256m: Limit the container to 256 MiB of RAM
		--cpus 0.5: Limit the container to 50% of one CPU core
		--pids-limit 128: Limit the container to 128 processes
		image: Use the lesson image; Docker picks the variant matching the host architecture
		sleep infinity: Keep the container alive until explicitly removed
	*/
	return []string{
		"create",
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
