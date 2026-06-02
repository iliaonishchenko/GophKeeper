package crypto

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDeriveKeyDeterministic(t *testing.T) {
	salt := []byte("0123456789abcdef")
	k1, err := DeriveKey("пароль", salt)
	require.NoError(t, err)
	k2, err := DeriveKey("пароль", salt)
	require.NoError(t, err)

	assert.Len(t, k1, keyLen)
	assert.Equal(t, k1, k2)
}

func TestDeriveKeyDiffersBySaltAndPassword(t *testing.T) {
	k1, err := DeriveKey("пароль", []byte("0123456789abcdef"))
	require.NoError(t, err)
	k2, err := DeriveKey("пароль", []byte("fedcba9876543210"))
	require.NoError(t, err)
	k3, err := DeriveKey("другой", []byte("0123456789abcdef"))
	require.NoError(t, err)

	assert.NotEqual(t, k1, k2)
	assert.NotEqual(t, k1, k3)
}

func TestNewSaltRandom(t *testing.T) {
	s1, err := NewSalt()
	require.NoError(t, err)
	s2, err := NewSalt()
	require.NoError(t, err)

	assert.Len(t, s1, saltLen)
	assert.False(t, bytes.Equal(s1, s2))
}

func TestEncryptDecryptRoundTrip(t *testing.T) {
	key, err := DeriveKey("мастер-пароль", []byte("0123456789abcdef"))
	require.NoError(t, err)

	plaintext := []byte("очень секретные данные")
	ciphertext, err := Encrypt(key, plaintext)
	require.NoError(t, err)
	assert.False(t, bytes.Equal(plaintext, ciphertext))

	got, err := Decrypt(key, ciphertext)
	require.NoError(t, err)
	assert.Equal(t, plaintext, got)
}

func TestDecryptWrongKey(t *testing.T) {
	key1, _ := DeriveKey("пароль1", []byte("0123456789abcdef"))
	key2, _ := DeriveKey("пароль2", []byte("0123456789abcdef"))

	ciphertext, err := Encrypt(key1, []byte("данные"))
	require.NoError(t, err)

	_, err = Decrypt(key2, ciphertext)
	assert.Error(t, err)
}

func TestDecryptTooShort(t *testing.T) {
	key, _ := DeriveKey("пароль", []byte("0123456789abcdef"))
	_, err := Decrypt(key, []byte("кор"))
	assert.ErrorIs(t, err, ErrCiphertextTooShort)
}
