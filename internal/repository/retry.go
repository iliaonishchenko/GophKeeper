package repository

import (
	"fmt"
	"time"
)

func executeWithRetry(classifier ErrorClassifier, operation func() error) error {
	const (
		maxAttempts = 4
		deltaDelay  = 2 * time.Second
		baseDelay   = 1 * time.Second
	)

	currDelay := baseDelay
	var lastErr error
	for attempt := 0; attempt < maxAttempts; attempt++ {
		if attempt != 0 {
			time.Sleep(currDelay)
			currDelay += deltaDelay
		}

		err := operation()
		if err == nil {
			return nil
		}

		if !classifier.IsRetriable(err) {
			return err
		}
		lastErr = err
	}
	return fmt.Errorf("операция не удалась после %d попыток: %w", maxAttempts, lastErr)
}
