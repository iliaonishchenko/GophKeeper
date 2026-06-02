package clientapp

import (
	"encoding/json"
	"fmt"

	"github.com/iliaonishchenko/gophkeeper/internal/crypto"
	"github.com/iliaonishchenko/gophkeeper/internal/model"
)

type CredentialsPayload struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type CardPayload struct {
	Number string `json:"number"`
	Holder string `json:"holder"`
	Expiry string `json:"expiry"`
	CVV    string `json:"cvv"`
}

func EncryptPayload(key []byte, typ model.ItemType, payload any) ([]byte, error) {
	plaintext, err := encodePayload(typ, payload)
	if err != nil {
		return nil, err
	}
	return crypto.Encrypt(key, plaintext)
}

func DecryptPayload(key []byte, ciphertext []byte) ([]byte, error) {
	return crypto.Decrypt(key, ciphertext)
}

func encodePayload(typ model.ItemType, payload any) ([]byte, error) {
	switch typ {
	case model.TypeText, model.TypeBinary:
		b, ok := payload.([]byte)
		if !ok {
			return nil, fmt.Errorf("ожидались байты для типа %d", typ)
		}
		return b, nil
	case model.TypeCredentials, model.TypeCard:
		return json.Marshal(payload)
	default:
		return nil, fmt.Errorf("неизвестный тип секрета: %d", typ)
	}
}
