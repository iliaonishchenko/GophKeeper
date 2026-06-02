package token

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIssueVerifyRoundTrip(t *testing.T) {
	const (
		secret = "super-secret"
		userID = "user-42"
	)

	tok := Issue(secret, userID, time.Hour)
	got, err := Verify(secret, tok)
	require.NoError(t, err)
	assert.Equal(t, userID, got)
}

func TestVerifyExpired(t *testing.T) {
	tok := Issue("secret", "user-1", -time.Minute)
	_, err := Verify("secret", tok)
	assert.ErrorIs(t, err, ErrExpiredToken)
}

func TestVerifyWrongSecret(t *testing.T) {
	tok := Issue("secret", "user-1", time.Hour)
	_, err := Verify("other-secret", tok)
	assert.ErrorIs(t, err, ErrInvalidToken)
}

func TestVerifyMalformed(t *testing.T) {
	tests := []struct {
		name string
		tok  string
	}{
		{name: "пустая строка", tok: ""},
		{name: "без подписи", tok: "payloadonly"},
		{name: "некорректный base64", tok: "!!!.signature"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Verify("secret", tt.tok)
			assert.Error(t, err)
		})
	}
}

func TestVerifyTamperedPayload(t *testing.T) {
	tok := Issue("secret", "user-1", time.Hour)

	_, sig, _ := splitToken(tok)
	_, err := Verify("secret", "dGFtcGVyZWQ."+sig)
	assert.ErrorIs(t, err, ErrInvalidToken)
}

func splitToken(tok string) (string, string, bool) {
	for i := 0; i < len(tok); i++ {
		if tok[i] == '.' {
			return tok[:i], tok[i+1:], true
		}
	}
	return "", "", false
}
