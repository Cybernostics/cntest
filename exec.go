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
	// db.AwaitLogPattern(30, "eady to")
	db.AwaitStartup(30)
	if c.StopAfterTest {
		defer func() {
			db.Stop(30)
			db.Remove()
		}()
	}

	userTestFn(t, c)
}
