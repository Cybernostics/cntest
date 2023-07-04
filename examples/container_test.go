package examples

import (
	"fmt"
	"testing"

	"github.com/corbym/gocrest/is"
	"github.com/cybernostics/cntest"
)

func TestContainer(t *testing.T) {
	cnt := cntest.NewContainer().WithImage("hello-world:latest")
	name, err := cnt.Start()
	defer cnt.Remove()
	assertThat(t, err, is.Nil())
	assertThat(t, len(name), is.GreaterThan(0))
	ok, err := cnt.AwaitExit(10)
	assertThat(t, err, is.Nil())
	assertThat(t, ok, is.True())
	logs, err := cnt.Logs()
	assertThat(t, err, is.Nil())
	assertThat(t, len(logs), is.GreaterThan(0))
	fmt.Printf("Logs %s\n", logs)

}
