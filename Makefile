.DEFAULT_GOAL := all

.PHONY: lint
lint:
	golangci-lint run ./...

.PHONY: all
all: lint test

.PHONY: test
test:
	go test -count 1 -p 1 -race ./...

