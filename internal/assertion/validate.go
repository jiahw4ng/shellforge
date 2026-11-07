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
func (check Assertion) Validate() error {
	switch check.Type {
	case AssertionTypeDirectoryExists, AssertionTypeFileExists:
		if strings.TrimSpace(check.Path) == "" {
			return fmt.Errorf(requiresPathErrorFormat, check.Type)
		}
	case AssertionTypeFileMode:
		if strings.TrimSpace(check.Path) == "" {
			return fmt.Errorf(requiresPathErrorFormat, check.Type)
		}
		if !fileMode.MatchString(check.Mode) {
			return fmt.Errorf(requiresModeErrorFormat, check.Type)
		}
	case AssertionTypeFileContent:
		if strings.TrimSpace(check.Path) == "" {
			return fmt.Errorf(requiresPathErrorFormat, check.Type)
		}
		if check.Contains == "" {
			return fmt.Errorf(requiresContentErrorFormat, check.Type)
		}
	case AssertionTypeCommandHistoryContains:
		if check.Contains == "" {
			return fmt.Errorf(requiresCommandErrorFormat, check.Type)
		}
	case AssertionTypeCurrentWorkingDirectory:
		if !path.IsAbs(check.Path) || path.Clean(check.Path) != check.Path {
			return fmt.Errorf(requiresAbsolutePathErrorFormat, check.Type)
		}
	case AssertionTypeEnvironmentVariableExists:
		if !environmentVariableName.MatchString(check.Name) {
			return fmt.Errorf(requiresEnvVarNameErrorFormat, check.Type)
		}
	default:
		return fmt.Errorf(unsupportedAssertionTypeErrorFormat, check.Type)
	}

	return nil
}
