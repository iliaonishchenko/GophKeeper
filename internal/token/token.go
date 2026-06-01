package token

import (
	"crypto/hmac"
	"encoding/base64"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/iliaonishchenko/gophkeeper/internal/signature"
)

var (
	ErrInvalidToken = errors.New("некорректный токен")
	ErrExpiredToken = errors.New("срок действия токена истёк")
)

func Issue(secret, userID string, ttl time.Duration) string {
	exp := time.Now().Add(ttl).Unix()
	raw := userID + "|" + strconv.FormatInt(exp, 10)
	payload := base64.RawURLEncoding.EncodeToString([]byte(raw))
	sig := signature.NewSignature(secret).Sign([]byte(payload))
	return payload + "." + sig
}

func Verify(secret, tok string) (string, error) {
	payload, sig, ok := strings.Cut(tok, ".")
	if !ok {
		return "", ErrInvalidToken
	}

	expected := signature.NewSignature(secret).Sign([]byte(payload))
	if !hmac.Equal([]byte(sig), []byte(expected)) {
		return "", ErrInvalidToken
	}

	raw, err := base64.RawURLEncoding.DecodeString(payload)
	if err != nil {
		return "", ErrInvalidToken
	}

	userID, expStr, ok := strings.Cut(string(raw), "|")
	if !ok || userID == "" {
		return "", ErrInvalidToken
	}

	exp, err := strconv.ParseInt(expStr, 10, 64)
	if err != nil {
		return "", ErrInvalidToken
	}
	if time.Now().Unix() > exp {
		return "", ErrExpiredToken
	}

	return userID, nil
}
