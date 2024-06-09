package wait

import (
	"errors"
	"time"
)

// UntilTrue call the fn every 500 millis until it errors, returns or a timeout
func UntilTrue(timeoutSeconds int, fn func() (bool, error)) (bool, error) {
	timeout := time.After(time.Duration(timeoutSeconds) * time.Second)
	tick := time.NewTicker(500 * time.Millisecond)
	// Keep trying until we're timed out or got a result or got an error
	for {
		select {
		// Got a timeout! fail with a timeout error
		case <-timeout:
			return false, errors.New("timed out")
		// Got a tick, we should check on doSomething()
		case <-tick.C:
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

// UntilDone Wait for the process to finish, error or timeout
func UntilDone(timeoutSeconds int, fn func() (chan bool, chan error)) (bool, error) {
	timeout := time.After(time.Duration(timeoutSeconds) * time.Second)
	doneChan, errChan := fn()
	// Keep trying until we're timed out or got a result or got an error
	for {
		select {
		// Got a timeout! fail with a timeout error
		case <-timeout:
			return false, errors.New("timed out")
		// Got a result
		case result := <-doneChan:
			return result, nil
		case err := <-errChan:
			return false, err
		}
	}
}
