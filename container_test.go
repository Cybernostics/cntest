package cntest_test

import (
	"fmt"
	"testing"

	"github.com/corbym/gocrest/is"
	"github.com/corbym/gocrest/then"
	"github.com/cybernostics/cntest"
)

func TestContainer(t *testing.T) {
	cntest.PullImage("hello-world", "latest", cntest.FromDockerHub)
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

func Test(t *testing.T) {
	cntest.PullImage("wiremock/wiremock", "latest", cntest.FromDockerHub)

	oldCont := cntest.FindContainer("wiremocky")
	if oldCont != nil {
		err := cntest.RemoveContainer(oldCont.Instance.ID)
		then.AssertThat(t, err, is.Nil())
	}

	cnt := cntest.NewContainer().WithImage("wiremock/wiremock")
	cnt.SetName("wiremocky")
	str8080 := fmt.Sprintf("%d", 8080)
	cnt.AddExposedPort(cntest.ContainerPort(str8080))
	cnt.AddPortMap(cntest.HostPort(str8080), cntest.ContainerPort(str8080))
	cnt.StopAfterTest = false
	cnt.RemoveAfterTest = false

	cntest.ExecuteWithRunningContainer(t, cnt, func(t *testing.T) {
		fmt.Print(cnt.Instance.ID)
		oldCont := cntest.FindContainer("wiremocky")
		then.AssertThat(t, oldCont, is.Not(is.Nil()))
		if oldCont != nil {
			cntest.RemoveContainer(oldCont.Instance.ID)
		}
	})

	oldCont = cntest.FindContainer("wiremocky")
	var nilContainer *cntest.Container
	then.AssertThat(t, oldCont, is.EqualTo(nilContainer))

}
