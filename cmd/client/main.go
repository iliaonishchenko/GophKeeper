package main

import (
	"cmp"
	"fmt"
	"os"
)

var (
	buildVersion string
	buildDate    string
	buildCommit  string
)

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(1)
	}

	cmd := os.Args[1]
	args := os.Args[2:]

	var err error
	switch cmd {
	case "register":
		err = cmdRegister(args)
	case "login":
		err = cmdLogin(args)
	case "add":
		err = cmdAdd(args)
	case "list":
		err = cmdList(args)
	case "get":
		err = cmdGet(args)
	case "delete":
		err = cmdDelete(args)
	case "sync":
		err = cmdSync(args)
	case "version":
		printVersion()
	default:
		usage()
		os.Exit(1)
	}

	if err != nil {
		fmt.Fprintln(os.Stderr, "ошибка:", err)
		os.Exit(1)
	}
}

func printVersion() {
	fmt.Println("Build version: " + cmp.Or(buildVersion, "N/A"))
	fmt.Println("Build date: " + cmp.Or(buildDate, "N/A"))
	fmt.Println("Build commit: " + cmp.Or(buildCommit, "N/A"))
}

func usage() {
	fmt.Fprintln(os.Stderr, `GophKeeper — менеджер секретов

Использование: gophkeeper-client <команда> [аргументы]

Команды:
  register  регистрация нового пользователя
  login     аутентификация пользователя
  add       добавление секрета (credentials|text|binary|card)
  list      список сохранённых секретов
  get       получение и расшифровка секрета по идентификатору
  delete    удаление секрета
  sync      синхронизация с сервером
  version   информация о версии и дате сборки`)
}
