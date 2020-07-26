package exmaples

import (
	"testing"

	"github.com/wjase/cntest"
)

// TestMain This magic function wraps all the other tests
// see https://golang.org/pkg/testing/ and the section on TestMain
func TestMain(m *testing.M) {
	// pull the image before you start testing so you don't blow your timeout
	cntest.PullImage("mysql", "latest", cntest.FromcntestHub)
	cntest.PullImage("postgres", "latest", cntest.FromcntestHub)
	cntest.PullImage("hello-world", "latest", cntest.FromcntestHub)
	m.Run()
}
