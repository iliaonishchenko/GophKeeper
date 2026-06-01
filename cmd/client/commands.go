package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"

	"github.com/iliaonishchenko/gophkeeper/internal/clientapp"
	"github.com/iliaonishchenko/gophkeeper/internal/model"
)

var errNotAuthenticated = errors.New("требуется вход: выполните register или login")

func cmdRegister(args []string) error {
	fs := flag.NewFlagSet("register", flag.ExitOnError)
	addr := fs.String("a", "localhost:3200", "адрес gRPC-сервера")
	login := fs.String("login", "", "логин пользователя")
	password := fs.String("password", "", "мастер-пароль")
	_ = fs.Parse(args)

	if *login == "" || *password == "" {
		return errors.New("укажите -login и -password")
	}

	c, err := clientapp.NewClient(*addr, "")
	if err != nil {
		return err
	}
	defer c.Close()

	token, encSalt, err := c.Register(context.Background(), *login, *password)
	if err != nil {
		return err
	}

	s := &clientapp.Session{Login: *login, Token: token, EncSalt: encSalt, ServerAddress: *addr}
	if err := s.Save(); err != nil {
		return err
	}
	fmt.Println("регистрация выполнена, сеанс сохранён")
	return nil
}

func cmdLogin(args []string) error {
	fs := flag.NewFlagSet("login", flag.ExitOnError)
	addr := fs.String("a", "localhost:3200", "адрес gRPC-сервера")
	login := fs.String("login", "", "логин пользователя")
	password := fs.String("password", "", "мастер-пароль")
	_ = fs.Parse(args)

	if *login == "" || *password == "" {
		return errors.New("укажите -login и -password")
	}

	c, err := clientapp.NewClient(*addr, "")
	if err != nil {
		return err
	}
	defer c.Close()

	token, encSalt, err := c.Login(context.Background(), *login, *password)
	if err != nil {
		return err
	}

	s := &clientapp.Session{Login: *login, Token: token, EncSalt: encSalt, ServerAddress: *addr}
	if err := s.Save(); err != nil {
		return err
	}
	fmt.Println("вход выполнен, сеанс сохранён")
	return nil
}

func cmdAdd(args []string) error {
	fs := flag.NewFlagSet("add", flag.ExitOnError)
	typ := fs.String("type", "", "тип секрета: credentials|text|binary|card")
	name := fs.String("name", "", "имя секрета")
	metadata := fs.String("metadata", "", "произвольные текстовые метаданные")
	password := fs.String("password", "", "мастер-пароль для шифрования")
	login := fs.String("login", "", "логин (для credentials)")
	secret := fs.String("secret", "", "пароль (для credentials)")
	text := fs.String("text", "", "текст (для text)")
	file := fs.String("file", "", "путь к файлу (для binary)")
	number := fs.String("number", "", "номер карты (для card)")
	holder := fs.String("holder", "", "держатель карты (для card)")
	expiry := fs.String("expiry", "", "срок действия (для card)")
	cvv := fs.String("cvv", "", "CVV (для card)")
	_ = fs.Parse(args)

	if *name == "" || *password == "" {
		return errors.New("укажите -name и -password")
	}

	itemType, payload, err := buildPayload(*typ, payloadInput{
		login: *login, secret: *secret, text: *text, file: *file,
		number: *number, holder: *holder, expiry: *expiry, cvv: *cvv,
	})
	if err != nil {
		return err
	}

	s, c, err := authedClient()
	if err != nil {
		return err
	}
	defer c.Close()

	key, err := s.DeriveKey(*password)
	if err != nil {
		return err
	}

	it := &model.Item{Type: itemType, Name: *name, Metadata: *metadata}
	saved, err := clientapp.Add(context.Background(), c, s, key, it, payload)
	if err != nil {
		return err
	}
	if err := s.Save(); err != nil {
		return err
	}
	fmt.Printf("секрет добавлен: id=%s\n", saved.ID)
	return nil
}

func cmdList(args []string) error {
	fs := flag.NewFlagSet("list", flag.ExitOnError)
	_ = fs.Parse(args)

	s, err := clientapp.LoadSession()
	if err != nil {
		return err
	}
	if !s.IsAuthenticated() {
		return errNotAuthenticated
	}

	items := s.ActiveItems()
	if len(items) == 0 {
		fmt.Println("нет сохранённых секретов (возможно, требуется sync)")
		return nil
	}
	for _, it := range items {
		fmt.Printf("id=%s\tтип=%s\tимя=%s\tметаданные=%s\n", it.ID, typeName(it.Type), it.Name, it.Metadata)
	}
	return nil
}

func cmdGet(args []string) error {
	fs := flag.NewFlagSet("get", flag.ExitOnError)
	id := fs.String("id", "", "идентификатор секрета")
	password := fs.String("password", "", "мастер-пароль для расшифровки")
	out := fs.String("out", "", "файл для сохранения бинарных данных")
	_ = fs.Parse(args)

	if *id == "" || *password == "" {
		return errors.New("укажите -id и -password")
	}

	s, c, err := authedClient()
	if err != nil {
		return err
	}
	defer c.Close()

	it := s.FindItem(*id)
	if it == nil {
		it, err = c.Get(context.Background(), *id)
		if err != nil {
			return err
		}
	}

	key, err := s.DeriveKey(*password)
	if err != nil {
		return err
	}
	plaintext, err := clientapp.Decrypt(key, it)
	if err != nil {
		return err
	}

	if it.Type == model.TypeBinary && *out != "" {
		if err := os.WriteFile(*out, plaintext, 0o600); err != nil {
			return err
		}
		fmt.Printf("бинарные данные сохранены в %s\n", *out)
		return nil
	}

	fmt.Printf("имя: %s\nтип: %s\nметаданные: %s\nданные: %s\n", it.Name, typeName(it.Type), it.Metadata, string(plaintext))
	return nil
}

func cmdDelete(args []string) error {
	fs := flag.NewFlagSet("delete", flag.ExitOnError)
	id := fs.String("id", "", "идентификатор секрета")
	_ = fs.Parse(args)

	if *id == "" {
		return errors.New("укажите -id")
	}

	s, c, err := authedClient()
	if err != nil {
		return err
	}
	defer c.Close()

	if err := c.Delete(context.Background(), *id); err != nil {
		return err
	}
	if _, err := clientapp.Sync(context.Background(), c, s); err != nil {
		return err
	}
	if err := s.Save(); err != nil {
		return err
	}
	fmt.Println("секрет удалён")
	return nil
}

func cmdSync(args []string) error {
	fs := flag.NewFlagSet("sync", flag.ExitOnError)
	_ = fs.Parse(args)

	s, c, err := authedClient()
	if err != nil {
		return err
	}
	defer c.Close()

	n, err := clientapp.Sync(context.Background(), c, s)
	if err != nil {
		return err
	}
	if err := s.Save(); err != nil {
		return err
	}
	fmt.Printf("синхронизировано изменений: %d\n", n)
	return nil
}

func authedClient() (*clientapp.Session, *clientapp.Client, error) {
	s, err := clientapp.LoadSession()
	if err != nil {
		return nil, nil, err
	}
	if !s.IsAuthenticated() {
		return nil, nil, errNotAuthenticated
	}
	c, err := clientapp.NewClient(s.ServerAddress, s.Token)
	if err != nil {
		return nil, nil, err
	}
	return s, c, nil
}

type payloadInput struct {
	login, secret  string
	text           string
	file           string
	number, holder string
	expiry, cvv    string
}

func buildPayload(typ string, in payloadInput) (model.ItemType, any, error) {
	switch typ {
	case "credentials":
		return model.TypeCredentials, clientapp.CredentialsPayload{Login: in.login, Password: in.secret}, nil
	case "text":
		return model.TypeText, []byte(in.text), nil
	case "binary":
		if in.file == "" {
			return 0, nil, errors.New("для binary укажите -file")
		}
		data, err := os.ReadFile(in.file)
		if err != nil {
			return 0, nil, fmt.Errorf("чтение файла: %w", err)
		}
		return model.TypeBinary, data, nil
	case "card":
		return model.TypeCard, clientapp.CardPayload{Number: in.number, Holder: in.holder, Expiry: in.expiry, CVV: in.cvv}, nil
	default:
		return 0, nil, errors.New("неизвестный тип: используйте credentials|text|binary|card")
	}
}

func typeName(t model.ItemType) string {
	switch t {
	case model.TypeCredentials:
		return "credentials"
	case model.TypeText:
		return "text"
	case model.TypeBinary:
		return "binary"
	case model.TypeCard:
		return "card"
	default:
		return "unspecified"
	}
}
