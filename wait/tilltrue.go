package wait

import (
	"errors"
	"time"
)

func UntilTrue(timeoutSeconds int, fn func() (bool, error)) (bool, error) {
	timeout := time.After(time.Duration(timeoutSeconds) * time.Second)
	tick := time.Tick(500 * time.Millisecond)
	// Keep trying until we're timed out or got a result or got an error
	for {
		select {
		// Got a timeout! fail with a timeout error
		case <-timeout:
			return false, errors.New("timed out")
		// Got a tick, we should check on doSomething()
		case <-tick:
			result, err := fn()
			if err != nil {
				return false, err
			}
			if result {
				return result, nil
			}
		}
	}

}
