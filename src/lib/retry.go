package lib

import (
	"math"
	"time"
)

// RetryOperation ...
func RetryOperation(operation func() error, maxRetries int) error {
	var err error
	for attempt := 0; attempt < maxRetries; attempt++ {
		err = operation()
		if err == nil {
			return nil
		}
		backoffTime := time.Duration(math.Exp2(float64(attempt))) * time.Second
		if backoffTime > 1*time.Minute {
			backoffTime = 1 * time.Minute
		}
		time.Sleep(backoffTime)
	}
	return err
}
