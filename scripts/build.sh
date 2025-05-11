#!/bin/bash

go generate ./api/pubsub
protoc --go_out=. --go-grpc_out=. api/pubsub/pubsub.proto

go build -o bin/server ./cmd/server

go build -o bin/example ./cmd/example