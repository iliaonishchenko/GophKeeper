package signature

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	correctKey   = "correct_key"
	incorrectKey = "incorrect_key"
	msg          = "test message"
)

func TestSign(t *testing.T) {
	tests := []struct {
		name        string
		key         string
		shouldMatch bool
	}{
		{
			name:        "Корректная подпись",
			key:         correctKey,
			shouldMatch: true,
		},
		{
			name:        "Некорректная подпись",
			key:         incorrectKey,
			shouldMatch: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msgBytes := []byte(msg)

			sig := NewSignature(tt.key)
			result := sig.Sign(msgBytes)

			h := hmac.New(sha256.New, []byte(correctKey))
			h.Write(msgBytes)
			expected := hex.EncodeToString(h.Sum(nil))

			assert.Equal(t, tt.shouldMatch, result == expected)
		})
	}
}

func TestKey(t *testing.T) {
	sig := NewSignature(correctKey)
	assert.Equal(t, correctKey, sig.Key())
}
