package clientapp

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/iliaonishchenko/gophkeeper/internal/model"
)

type Session struct {
	Login         string        `json:"login"`
	Token         string        `json:"token"`
	EncSalt       []byte        `json:"enc_salt"`
	ServerAddress string        `json:"server_address"`
	Cursor        int64         `json:"cursor"`
	Items         []*model.Item `json:"items"`
}

func sessionDir() (string, error) {
	base, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("определение каталога конфигурации: %w", err)
	}
	return filepath.Join(base, "gophkeeper"), nil
}

func sessionPath() (string, error) {
	dir, err := sessionDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "session.json"), nil
}

func LoadSession() (*Session, error) {
	path, err := sessionPath()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return &Session{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("чтение сеанса: %w", err)
	}
	var s Session
	if err := json.Unmarshal(data, &s); err != nil {
		return nil, fmt.Errorf("разбор сеанса: %w", err)
	}
	return &s, nil
}

func (s *Session) Save() error {
	dir, err := sessionDir()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return fmt.Errorf("создание каталога конфигурации: %w", err)
	}
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return fmt.Errorf("сериализация сеанса: %w", err)
	}
	path := filepath.Join(dir, "session.json")
	if err := os.WriteFile(path, data, 0o600); err != nil {
		return fmt.Errorf("запись сеанса: %w", err)
	}
	return nil
}

func (s *Session) IsAuthenticated() bool {
	return s.Token != ""
}

func (s *Session) mergeItem(incoming *model.Item) {
	for i, existing := range s.Items {
		if existing.ID == incoming.ID {
			if incoming.Version >= existing.Version {
				s.Items[i] = incoming
			}
			return
		}
	}
	s.Items = append(s.Items, incoming)
}
