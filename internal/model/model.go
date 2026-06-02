package model

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
)

type ItemType int16

const (
	TypeUnspecified ItemType = 0
	TypeCredentials ItemType = 1
	TypeText        ItemType = 2
	TypeBinary      ItemType = 3
	TypeCard        ItemType = 4
)

type User struct {
	ID       string
	Login    string
	AuthHash []byte
	AuthSalt []byte
	EncSalt  []byte
}

type Item struct {
	ID         string
	UserID     string
	Type       ItemType
	Name       string
	Ciphertext []byte
	Metadata   string
	Version    int64
	UpdatedAt  int64
	Deleted    bool
}

func NewID() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("генерация идентификатора: %w", err)
	}
	return hex.EncodeToString(b), nil
}
