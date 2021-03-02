.PHONY: build run test vet

build:
	go build -o goproxy cmd/goproxy/main.go

run:
	go run cmd/goproxy/main.go config.json

test:
	go test ./...

vet:
	go vet ./...
