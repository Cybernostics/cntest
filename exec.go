package dbcontainers

import (
	"testing"
)

// DBTestFn implement your DB tests using this signature
type DBTestFn func(t *testing.T, config *DBContainer)

// ExecuteWithRunningDB wraps a test function by creating a db
func ExecuteWithRunningDB(t *testing.T, c *DBContainer, userTestFn DBTestFn) {
	db, err := c.Start()
	if err != nil {
		t.Fatalf("Couldn't start container %v", err)
	}
	db.AwaitIsRunning(c.MaxStartTimeSeconds)
	if c.StopAfterTest {
		defer func() {
			_, err := db.Stop(10)
			if err != nil {
				t.Errorf("Couldn't stop container: %s\n Error was %v", db.Instance.ID, err)
			}

			if c.RemoveAfterTest {
				err = db.Remove()
				if err != nil {
					t.Errorf("Couldn't stop container: %s\n Error was %v", db.Instance.ID, err)
				}
			}
		}()
	}

	userTestFn(t, c)
}
