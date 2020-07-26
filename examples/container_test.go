package examples

import (
	"fmt"
	"testing"

	"github.com/corbym/gocrest/is"
	"github.com/corbym/gocrest/then"
	"github.com/cybernostics/cntest"
)

func TestContainer(t *testing.T) {
	cnt := cntest.NewContainer().WithImage("hello-world")
	name, err := cnt.Start()
	defer cnt.Remove()
	then.AssertThat(t, err, is.Nil())
	then.AssertThat(t, len(name), is.GreaterThan(0))
	ok, err := cnt.AwaitExit(10)
	then.AssertThat(t, err, is.Nil())
	then.AssertThat(t, ok, is.True())
	logs, err := cnt.Logs()
	then.AssertThat(t, err, is.Nil())
	then.AssertThat(t, len(logs), is.GreaterThan(0))
	fmt.Printf("Logs %s\n", logs)

}
