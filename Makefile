.PHONY: build fmt run test tests vet

build:
	go build -o goproxy cmd/goproxy/main.go

fmt:
	go fmt ./...

run:
	go run cmd/goproxy/main.go config.json

test:
	go test -race -covermode=atomic ./...

tests: fmt vet test

vet:
	go vet ./...
