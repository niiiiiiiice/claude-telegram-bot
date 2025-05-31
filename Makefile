.PHONY: build run docker-build docker-run test clean

build:
	go build -o bin/main.exe cmd/main.go

run:
	go run cmd/main.go

docker-build:
	docker build -t telegram-chatbot .

docker-run:
	docker run --env-file .env telegram-chatbot

test:
	go test ./...

clean:
	rm -rf bin/

wire:
	cd internal/di && wire

# Для генерации wire кода
generate:
	go generate ./...