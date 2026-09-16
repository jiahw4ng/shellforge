package container

import "sync"

// LessonContainer is one disposable, isolated lesson environment.
type LessonContainer struct {
	Name string
	// RemoveOnce ensures that Remove() is only called once per container.
	RemoveOnce sync.Once
}
