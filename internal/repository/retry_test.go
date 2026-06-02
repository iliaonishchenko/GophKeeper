package repository

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeClassifier struct {
	retriable bool
}

func (f fakeClassifier) IsRetriable(error) bool {
	return f.retriable
}

func TestExecuteWithRetrySuccess(t *testing.T) {
	classifier := fakeClassifier{retriable: false}
	calls := 0
	err := executeWithRetry(classifier, func() error {
		calls++
		return nil
	})
	require.NoError(t, err)
	assert.Equal(t, 1, calls)
}

func TestExecuteWithRetryNonRetriable(t *testing.T) {
	classifier := fakeClassifier{retriable: false}
	calls := 0
	err := executeWithRetry(classifier, func() error {
		calls++
		return errors.New("фатальная ошибка")
	})
	assert.Error(t, err)
	assert.Equal(t, 1, calls, "нерекуррентная ошибка не должна повторяться")
}

func TestExecuteWithRetryRetriableExhausted(t *testing.T) {
	classifier := fakeClassifier{retriable: true}
	calls := 0
	err := executeWithRetry(classifier, func() error {
		calls++
		return errors.New("временная ошибка")
	})
	assert.Error(t, err)
	assert.Equal(t, 4, calls, "ретраебл-ошибка должна исчерпать все попытки")
}
