.PHONY: run build migrate

run:
	go run cmd/server/main.go

build:
	go build -o bin/server cmd/server/main.go

migrate:
	@echo "Migrations run automatically on server start"

tidy:
	go mod tidy
