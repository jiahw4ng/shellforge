package assertion

const (
	// $1 is the path to check for existence
	DirectoryExistsScript = `if [ -d "$1" ]; then printf present; else printf missing; fi`
	// $1 is the path to check for existence
	FileExistsScript = `if [ -f "$1" ]; then printf present; else printf missing; fi`
	// $1 is the path to check for existence
	// $2 is the content to check for
	FileContentScript = `if [ -f "$1" ] && grep -Fq -- "$2" "$1"; then printf present; else printf missing; fi`
	// $1 is the Bash history file to check
	// $2 is the command text to find
	CommandHistoryContainsScript = `if [ -f "$1" ] && grep -Fq -- "$2" "$1"; then printf present; else printf missing; fi`
)
