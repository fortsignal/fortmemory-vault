.PHONY: build test vet fmt tidy install run-help

build:
	go build -o bin/fortmemory ./cmd/fortmemory

test:
	go test ./...

vet:
	go vet ./...

fmt:
	gofmt -w ./cmd ./internal

tidy:
	go mod tidy

install:
	go install ./cmd/fortmemory

run-help: build
	./bin/fortmemory help

# Keep dashboard embed in sync with web source
sync-ui:
	cp web/index.html internal/server/static/index.html
