package assertion

type AssertionKind string

const (
	AssertionTypeDirectoryExists AssertionKind = "directory_exists"
	AssertionTypeFileExists      AssertionKind = "file_exists"
	AssertionTypeFileContains    AssertionKind = "file_contains"
)

// Assertion describes one future state check run separately inside the lesson
// container. It verifies results rather than requiring one exact command.
type Assertion struct {
	Type     AssertionKind `yaml:"type"`
	Path     string        `yaml:"path,omitempty"`
	Contains string        `yaml:"contains,omitempty"`
	Mode     string        `yaml:"mode,omitempty"`
}

// Result is the outcome of one assertion check run inside the lesson container.
type Result struct {
	Assertion Assertion
	Passed    bool
	Message   string
}
