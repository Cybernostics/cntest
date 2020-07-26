package cntest

import (
	"fmt"
	"testing"
)

// DBTestFn implement your DB tests using this signature
type DBTestFn func(t *testing.T, config *Container)

// ExecuteWithRunningDB wraps a test function by creating a db
func ExecuteWithRunningDB(t *testing.T, c *Container, userTestFn DBTestFn) {
	isOk := false
	containerID, err := c.Start()
	defer func() {
		if err := recover(); err != nil {
			fmt.Printf("Panic within ExecuteWithRunningDB. Error was: %v", err)
			c.Stop(0)
			c.Remove()
		}
	}()
	defer func() {
		if !isOk {
			t.Errorf("Failed to run container.")
			logsStr, err := c.Logs()
			if err != nil {
				fmt.Printf("Logs were: %s\n", logsStr)
			}
			c.Stop(0)
			c.Remove()
		}
	}()
	if err != nil {
		t.Fatalf("Couldn't start container %v", err)
		return
	}
	fmt.Printf("Started %s\n", containerID)
	if ok, err := c.AwaitIsReady(); !ok {
		defer c.Stop(10)
		defer c.Remove()
		if err != nil {
			t.Errorf("Couldn't start container: %s\n Error was %v", c.Instance.ID, err)
		}
	}
	if c.IsStopAfterTest() {
		defer func() {
			_, err := c.Stop(10)
			if err != nil {
				t.Errorf("Couldn't stop container: %s\n Error was %v", c.Instance.ID, err)
			}

			if c.IsRemoveAfterTest() {
				err = c.Remove()
				if err != nil {
					t.Errorf("Couldn't stop container: %s\n Error was %v", c.Instance.ID, err)
				}
			}
		}()
	}
	userTestFn(t, c)
	isOk = true
}
