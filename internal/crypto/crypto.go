package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/pbkdf2"
	"crypto/rand"
	"crypto/sha256"
	"errors"
	"fmt"
)

const (
	keyLen     = 32
	iterations = 600000
	saltLen    = 16
)

var ErrCiphertextTooShort = errors.New("шифртекст слишком короткий")

func NewSalt() ([]byte, error) {
	salt := make([]byte, saltLen)
	if _, err := rand.Read(salt); err != nil {
		return nil, fmt.Errorf("генерация соли: %w", err)
	}
	return salt, nil
}

func DeriveKey(password string, salt []byte) ([]byte, error) {
	key, err := pbkdf2.Key(sha256.New, password, salt, iterations, keyLen)
	if err != nil {
		return nil, fmt.Errorf("деривация ключа: %w", err)
	}
	return key, nil
}

func Encrypt(key, plaintext []byte) ([]byte, error) {
	gcm, err := newGCM(key)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return nil, fmt.Errorf("сгенерировать nonce: %w", err)
	}

	return gcm.Seal(nonce, nonce, plaintext, nil), nil
}

func Decrypt(key, body []byte) ([]byte, error) {
	gcm, err := newGCM(key)
	if err != nil {
		return nil, err
	}

	if len(body) < gcm.NonceSize() {
		return nil, ErrCiphertextTooShort
	}

	nonce := body[:gcm.NonceSize()]
	ciphertext := body[gcm.NonceSize():]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("aes-gcm open: %w", err)
	}
	return plaintext, nil
}

func newGCM(key []byte) (cipher.AEAD, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("новый aes шифр: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("новый gcm: %w", err)
	}
	return gcm, nil
}
