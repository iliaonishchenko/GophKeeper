package repository

import (
	"fmt"
	"time"
)

func executeWithRetry(classifier *PostgresErrorClassifier, operation func() error) error {
	const (
		maxAttempts = 4
		deltaDelay  = 2 * time.Second
	)
	currDelay := 1 * time.Second
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
