package container

// Step names one stage of starting a lesson sandbox. Steps are identifiers only:
// the wording a learner sees belongs to the screen that renders them.
type Step int

const (
	// StepCheckingDocker means the Docker CLI and engine are being checked.
	StepCheckingDocker Step = iota + 1
	// StepCheckingImage means the lesson image is being looked up locally.
	StepCheckingImage
	// StepPullingImage means the image was missing and is being downloaded.
	StepPullingImage
	// StepStartingContainer means the lesson container is being created and started.
	StepStartingContainer
	// StepPreparingLesson means the lesson's setup commands are running.
	StepPreparingLesson
	// StepConnectingShell means the learner's Bash session is being attached.
	StepConnectingShell
)

// ProgressFunc is told which Step startup has reached. A nil ProgressFunc is
// valid and ignores every step.
type ProgressFunc func(Step)

// Report announces that startup has reached step.
func (f ProgressFunc) Report(step Step) {
	if f != nil {
		f(step)
	}
}
