package assertion

import (
	"fmt"
	"path"
	"regexp"
	"strings"
)

var (
	environmentVariableName = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)
	fileMode                = regexp.MustCompile(`^[0-7]{3}$`)
)

const (
	requiresPathErrorFormat             = "%s requires a path"
	requiresModeErrorFormat             = "%s requires a three-digit octal mode"
	requiresContentErrorFormat          = "%s requires content to find"
	requiresCommandErrorFormat          = "%s requires command text to find"
	requiresAbsolutePathErrorFormat     = "%s requires a normalized absolute path"
	requiresEnvVarNameErrorFormat       = "%s requires a valid environment variable name"
	unsupportedAssertionTypeErrorFormat = "unsupported assertion type %q"
)

// Validate rejects unsupported assertion types and incomplete assertion data
// before a lesson is embedded in the application binary.
func (a Assertion) Validate() error {
	switch a.Type {
	case AssertionTypeDirectoryExists, AssertionTypeFileExists:
		if strings.TrimSpace(a.Path) == "" {
			return fmt.Errorf(requiresPathErrorFormat, a.Type)
		}
	case AssertionTypeFileMode:
		if strings.TrimSpace(a.Path) == "" {
			return fmt.Errorf(requiresPathErrorFormat, a.Type)
		}
		if !fileMode.MatchString(a.Mode) {
			return fmt.Errorf(requiresModeErrorFormat, a.Type)
		}
	case AssertionTypeFileContent:
		if strings.TrimSpace(a.Path) == "" {
			return fmt.Errorf(requiresPathErrorFormat, a.Type)
		}
		if a.Contains == "" {
			return fmt.Errorf(requiresContentErrorFormat, a.Type)
		}
	case AssertionTypeCommandHistoryContains:
		if a.Contains == "" {
			return fmt.Errorf(requiresCommandErrorFormat, a.Type)
		}
	case AssertionTypeCurrentWorkingDirectory:
		if !path.IsAbs(a.Path) || path.Clean(a.Path) != a.Path {
			return fmt.Errorf(requiresAbsolutePathErrorFormat, a.Type)
		}
	case AssertionTypeEnvironmentVariableExists:
		if !environmentVariableName.MatchString(a.Name) {
			return fmt.Errorf(requiresEnvVarNameErrorFormat, a.Type)
		}
	default:
		return fmt.Errorf(unsupportedAssertionTypeErrorFormat, a.Type)
	}

	return nil
}
