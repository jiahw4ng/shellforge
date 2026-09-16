package container

import "sync"

// Container is one disposable, isolated lesson environment.
type Container struct {
	Name string
	// RemoveOnce ensures that Remove() is only called once per container.
	RemoveOnce sync.Once
}
