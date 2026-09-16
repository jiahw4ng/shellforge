package container

import "sync"

// LessonContainer is one disposable, isolated lesson environment.
type LessonContainer struct {
	name string
	// removeOnce ensures that Remove() is only called once per container.
	removeOnce sync.Once
}
