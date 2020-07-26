package cntest

import (
	"testing"
)

// ContainerMap is a type alias for a collection of containers
type ContainerMap map[string]*Container

// ContainerTestFn implement your container tests using this signature
type ContainerTestFn func(t *testing.T, containers ContainerGroup)

// ExecuteWithRunningContainer wraps a test function by creating a container
// for you to test and ensures it gets cleaned up in hte end
func ExecuteWithRunningContainer(t *testing.T, cnts ContainerGroup, userTestFn ContainerTestFn) {
	cnts.Start()

	cnts.Await()

	userTestFn(t, cnts)
	// if db.IsStopAfterTest() {
	// 	defer func() {
	// 		_, err := db.Stop(10)
	// 		if err != nil {
	// 			t.Errorf("Couldn't stop container: %s\n Error was %v", db.Instance.ID, err)
	// 		}

	// 		if c.IsRemoveAfterTest() {
	// 			err = db.Remove()
	// 			if err != nil {
	// 				t.Errorf("Couldn't stop container: %s\n Error was %v", db.Instance.ID, err)
	// 			}
	// 		}
	// 	}()
	// }

}
