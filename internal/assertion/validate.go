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

// Validate rejects unsupported assertion types and incomplete assertion data
// before a lesson is embedded in the application binary.
func (check Assertion) Validate() error {
	switch check.Type {
	case AssertionTypeDirectoryExists, AssertionTypeFileExists:
		if strings.TrimSpace(check.Path) == "" {
			return fmt.Errorf("%s requires a path", check.Type)
		}
	case AssertionTypeFileMode:
		if strings.TrimSpace(check.Path) == "" {
			return fmt.Errorf("%s requires a path", check.Type)
		}
		if !fileMode.MatchString(check.Mode) {
			return fmt.Errorf("%s requires a three-digit octal mode", check.Type)
		}
	case AssertionTypeFileContent:
		if strings.TrimSpace(check.Path) == "" {
			return fmt.Errorf("%s requires a path", check.Type)
		}
		if check.Contains == "" {
			return fmt.Errorf("%s requires content to find", check.Type)
		}
	case AssertionTypeCommandHistoryContains:
		if check.Contains == "" {
			return fmt.Errorf("%s requires command text to find", check.Type)
		}
	case AssertionTypeCurrentWorkingDirectory:
		if !path.IsAbs(check.Path) || path.Clean(check.Path) != check.Path {
			return fmt.Errorf("%s requires a normalized absolute path", check.Type)
		}
	case AssertionTypeEnvironmentVariableExists:
		if !environmentVariableName.MatchString(check.Name) {
			return fmt.Errorf("%s requires a valid environment variable name", check.Type)
		}
	default:
		return fmt.Errorf("unsupported assertion type %q", check.Type)
	}

	return nil
}
