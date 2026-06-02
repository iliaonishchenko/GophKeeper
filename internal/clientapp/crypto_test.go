package clientapp

import (
	"encoding/json"
	"testing"

	"github.com/iliaonishchenko/gophkeeper/internal/crypto"
	"github.com/iliaonishchenko/gophkeeper/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testKey(t *testing.T) []byte {
	t.Helper()
	key, err := crypto.DeriveKey("мастер", []byte("0123456789abcdef"))
	require.NoError(t, err)
	return key
}

func TestEncryptDecryptCredentials(t *testing.T) {
	key := testKey(t)
	payload := CredentialsPayload{Login: "alice", Password: "s3cret"}

	ct, err := EncryptPayload(key, model.TypeCredentials, payload)
	require.NoError(t, err)

	plain, err := DecryptPayload(key, ct)
	require.NoError(t, err)

	var got CredentialsPayload
	require.NoError(t, json.Unmarshal(plain, &got))
	assert.Equal(t, payload, got)
}

func TestEncryptDecryptCard(t *testing.T) {
	key := testKey(t)
	payload := CardPayload{Number: "4111111111111111", Holder: "ALICE", Expiry: "12/30", CVV: "123"}

	ct, err := EncryptPayload(key, model.TypeCard, payload)
	require.NoError(t, err)

	plain, err := DecryptPayload(key, ct)
	require.NoError(t, err)

	var got CardPayload
	require.NoError(t, json.Unmarshal(plain, &got))
	assert.Equal(t, payload, got)
}

func TestEncryptDecryptText(t *testing.T) {
	key := testKey(t)
	ct, err := EncryptPayload(key, model.TypeText, []byte("секретная заметка"))
	require.NoError(t, err)

	plain, err := DecryptPayload(key, ct)
	require.NoError(t, err)
	assert.Equal(t, "секретная заметка", string(plain))
}

func TestEncryptDecryptBinary(t *testing.T) {
	key := testKey(t)
	data := []byte{0x00, 0x01, 0x02, 0xff}
	ct, err := EncryptPayload(key, model.TypeBinary, data)
	require.NoError(t, err)

	plain, err := DecryptPayload(key, ct)
	require.NoError(t, err)
	assert.Equal(t, data, plain)
}

func TestEncryptWrongPayloadType(t *testing.T) {
	key := testKey(t)
	_, err := EncryptPayload(key, model.TypeText, "не байты")
	assert.Error(t, err)
}

func TestEncryptUnknownType(t *testing.T) {
	key := testKey(t)
	_, err := EncryptPayload(key, model.TypeUnspecified, []byte("x"))
	assert.Error(t, err)
}
