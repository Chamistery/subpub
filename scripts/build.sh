#!/usr/bin/env bash

# Генерация protobuf кода
protoc \
  -I. \
  --go_out=. --go_opt=paths=source_relative \
  --go-grpc_out=. --go-grpc_opt=paths=source_relative \
  api/pubsub/pubsub.proto

# Сборка бинарников
mkdir -p bin

go mod tidy
go build -o bin/server ./cmd/server
go build -o bin/example ./cmd/example