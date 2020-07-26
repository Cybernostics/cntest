package examples

import (
	"testing"

	"github.com/cybernostics/cntest"
)

// TestMain This magic function wraps all the other tests
// see https://golang.org/pkg/testing/ and the section on TestMain
func TestMain(m *testing.M) {
	// pull the image before you start testing so you don't blow your timeout
	cntest.PullImage("mysql", "latest", cntest.FromDockerHub)
	cntest.PullImage("postgres", "latest", cntest.FromDockerHub)
	cntest.PullImage("hello-world", "latest", cntest.FromDockerHub)
	m.Run()
}
