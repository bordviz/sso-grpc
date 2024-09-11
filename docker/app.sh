#!/bin/bash

go run cmd/migrations/main.go --action=up

go build -o build/main cmd/sso/main.go

./build/main