package assertion

type AssertionKind string

const (
	// checks whether a directory exists at the given path
	AssertionTypeDirectoryExists AssertionKind = "directory_exists"
	// checks whether a file exists at the given path
	AssertionTypeFileExists AssertionKind = "file_exists"
	// checks whether a file exists at the given path and contains the given string
	AssertionTypeFileContent AssertionKind = "file_content"
	// checks whether a regular file has the given three-digit octal permission mode
	AssertionTypeFileMode AssertionKind = "file_mode"
	// checks whether the command history contains the given string
	AssertionTypeCommandHistoryContains AssertionKind = "command_history_contains"
	// checks whether the learner's latest interactive working directory matches the given path
	AssertionTypeCurrentWorkingDirectory AssertionKind = "cwd"
	// checks whether the learner's latest interactive environment exports the given variable
	AssertionTypeEnvironmentVariableExists AssertionKind = "environment_variable_exists"
)

const (
	// $1 is the path to check for existence
	directoryExistsScript = `if [ -d "$1" ]; then printf present; else printf missing; fi`
	// $1 is the path to check for existence
	fileExistsScript = `if [ -f "$1" ]; then printf present; else printf missing; fi`
	// $1 is the regular file to inspect and $2 is its expected octal mode.
	fileModeScript = `if [ -f "$1" ] && [ "$(stat -c '%a' -- "$1")" = "$2" ]; then printf present; else printf missing; fi`
	// $1 is the path to check for existence
	// $2 is the content to check for
	fileContentScript = `if [ -f "$1" ] && grep -Fq -- "$2" "$1"; then printf present; else printf missing; fi`
	// $1 is the Bash history file to check
	// $2 is the command text to find
	commandHistoryContainsScript = `if [ -f "$1" ] && grep -Fq -- "$2" "$1"; then printf present; else printf missing; fi`
	// $1 records the learner's physical working directory at every interactive prompt.
	// $2 is the expected normalized absolute path.
	currentWorkingDirectoryScript = `if [ -f "$1" ] && [ "$(tail -n 1 -- "$1")" = "$2" ]; then printf present; else printf missing; fi`
	// $1 is an env -0 snapshot and $2 is an expected environment variable name.
	// Read each NUL-delimited entry independently so a matching string inside a
	// different variable's value cannot pass the assertion.
	environmentVariableExistsScript = `if [ ! -f "$1" ]; then
	printf missing
	exit
	fi
	while IFS= read -r -d '' entry; do
	case "$entry" in
	"$2"=*) printf present; exit ;;
	esac
	done < "$1"
	printf missing`
)

const (
	learnerWorkspace            = "/home/student/workspace"
	commandHistoryFile          = "/home/student/.shellforge-history"
	workingDirectoryHistoryFile = "/home/student/.shellforge-working-directories"
	environmentSnapshotFile     = "/home/student/.shellforge-environment"
)
