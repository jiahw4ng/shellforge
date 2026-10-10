package container

import "sync"

// Container is one disposable, isolated lesson environment.
type Container struct {
	name string
	// removeOnce ensures that Remove() is only called once per container.
	removeOnce sync.Once
}
