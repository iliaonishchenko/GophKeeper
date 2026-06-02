## Сборка

```sh
make proto    # генерация gRPC-кода
make mocks    # генерация моков для тестов
make build    # сборка bin/gophkeeper-server и bin/gophkeeper-client
make test     # юнит-тесты
make cover    # покрытие тестами
```

Версия и дата сборки внедряются через ldflags (`make build`):

```sh
./bin/gophkeeper-client version
```

## Запуск сервера

```sh
DATABASE_DSN="postgres://user:pass@localhost:5432/gophkeeper?sslmode=disable" \
TOKEN_SECRET="секрет" \
./bin/gophkeeper-server -g localhost:3200
```

## Использование клиента

```sh
# регистрация / вход
gophkeeper-client register -a localhost:3200 -login alice -password master
gophkeeper-client login    -a localhost:3200 -login alice -password master

# добавление секретов разных типов
gophkeeper-client add -type credentials -name "сайт" -login alice@site -secret p@ss -password master
gophkeeper-client add -type text   -name "заметка" -text "секрет" -password master
gophkeeper-client add -type binary -name "ключ"    -file ./key.bin -password master
gophkeeper-client add -type card   -name "виза" -number 4111... -holder ALICE -expiry 12/30 -cvv 123 -password master

# просмотр и синхронизация
gophkeeper-client sync
gophkeeper-client list
gophkeeper-client get    -id <id> -password master
gophkeeper-client delete -id <id>
```
