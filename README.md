# SubPub

Выполнил **Камлев Виталий**

Простой Pub-Sub сервис на Go с gRPC интерфейсом. Использованы паттерны:
- *Observer* (Publish–Subscribe): пускаются события (Publish), а подписчики их получают через свои очереди (Subscribe)
- *Dependency Injection*: в организации сервисов инициализация через конструкторы (NewSubPub, NewServer(bus, log)) и передача зависимостей в обработчики
- *Graceful ShutDown*: сервис корректно завершает работу по сигналам `SIGINT` и `SIGTERM`, дожидаясь завершения доставки сообщений и остановки gRPC-сервера.

## Структура проекта

```
subpub/
├── cmd/
│   ├── example/       # CLI-пример использования pkg/subpub
│   └── server/        # gRPC-сервис
├── api/pubsub/        # protobuf-схемы и сгенерированный код
├── pkg/subpub/        # библиотечный пакет Pub-Sub
├── internal/          # внутренняя логика сервиса
│   ├── config/        # загрузка и структура конфига
│   ├── logger/        # инициализация логгера
│   └── service/       # реализация gRPC-сервиса
├── configs/           # файлы конфигурации (YAML)
├── scripts/           # вспомогательные скрипты сборки
├── go.mod             # зависимости модуля
├── go.sum             # контроль версий зависимостей
└── README.md          # этот файл
```

## Требования

- Go 1.21+
- protoc (Protocol Buffers Compiler)
- Плагины protoc-gen-go и protoc-gen-go-grpc

## Установка зависимостей

```bash
go mod download
```

## Генерация protobuf-кода

```bash
# из корня проекта
protoc --go_out=. --go-grpc_out=. api/pubsub/pubsub.proto
```

## Сборка

```bash
# или использовать скрипт
chmod +x scripts/build.sh
scripts/build.sh
```

Файлы бинарников будут созданы в `bin/`: `server` и `example`.

## Запуск gRPC-сервера

```bash
# по умолчанию читает конфиг configs/dev.yaml
./bin/server
```

По умолчанию сервер слушает порт, указанный в `configs/dev.yaml` (например, `:50051`).

### Переменные окружения

Вы можете переопределить настройки конфига через ENV-переменные:

- `PORT` — порт для прослушивания, например `:60000`
- `SHUTDOWN_TIMEOUT` — таймаут graceful shutdown (строка в формате `5s`, `1m` и т.д.)

## Запуск примера клиента

```bash
go run cmd/example/main.go
```

Пример продемонстрирует подписку на тему `greeting` и публикацию нескольких сообщений.

## Демонстрация медленного подписчика

```bash
go run examples/slow_subscriber/main.go
```

В консоли будет видно, что быстрый подписчик обрабатывает сообщения без задержек медленного.

## Тестирование

### Unit-тесты

- Пакет `pkg/subpub`:
  ```bash
    go test ./pkg/subpub -v
  ```

- Внутренний сервис `internal/service`:
  ```bash
    go test ./internal/service -v
  ```

- Все тесты сразу:
  ```bash
    go test ./... -v
  ```

### Покрытие тестами

```bash
    go test ./... -coverprofile=coverage.out
    go tool cover -html=coverage.out
```

## Логирование

Используется `go.uber.org/zap` в режиме Production. Уровень логирования можно менять в коде при инициализации логгера.
