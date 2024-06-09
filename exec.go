package cntest

import (
	"fmt"
	"strings"
	"testing"
)

// ContainerTestFn implement your DB tests using this signature
type ContainerTestFn func(t *testing.T)

// ExecuteWithRunningContainer wraps a test function by creating a db
func ExecuteWithRunningContainer(t *testing.T, c *Container, userTestFn ContainerTestFn) {
	isOk := false
	containerID, err := c.Start()
	defer func() {
		if !isOk {
			t.Errorf("Failed to run container.")
			logsStr, err := c.Logs()
			if err != nil {
				fmt.Printf("Logs were: %s\n", logsStr)
			}
			if c.StopAfterTest {
				_, _ = c.Stop(0)
			}
			if c.RemoveAfterTest {
				_ = c.Remove()
			}
		}
	}()
	defer func() {
		if err := recover(); err != nil {
			fmt.Printf("Panic within ExecuteWithRunningContainer. Error was: %v", err)
			if c.StopAfterTest {
				_, _ = c.Stop(0)
			}
			if c.RemoveAfterTest {
				_ = c.Remove()
			}
		}
	}()
	if err != nil {
		t.Fatalf("Couldn't start container %v", err)
		return
	}
	fmt.Printf("Started %s - Awaiting ready\n", containerID)
	if ok, err := c.AwaitIsReady(); !ok {
		if c.StopAfterTest {
			defer func() {
				_, _ = c.Stop(10)
			}()
		}
		if c.RemoveAfterTest {
			defer func() {
				_ = c.Remove()
			}()
		}
		if err != nil {
			t.Errorf("Couldn't start container: %s\n Error was %v", c.Instance.ID, err)
		}
	}
	if c.IsStopAfterTest() {
		defer func() {
			_, err := c.Stop(10)
			if err != nil {
				if !strings.Contains(err.Error(),"No such container"){
					t.Errorf("Couldn't stop container: %s\n Error was %v", c.Instance.ID, err)
				}
			}

			if c.IsRemoveAfterTest() {
				err = c.Remove()
				if err != nil {
					if !strings.Contains(err.Error(),"No such container"){
						t.Errorf("Couldn't remove container: %s\n Error was %v", c.Instance.ID, err)
					}
					}
			}
		}()
	}
	if ready, err := c.ContainerReady(); err == nil && ready {
		userTestFn(t)
		isOk = true

	}
}
