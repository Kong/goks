.DEFAULT_GOAL := all


.PHONY: gen-lua-tree
gen-lua-tree:
	docker build -t goks -f build/Dockerfile .
	docker run -it -v $(PWD):/goks goks

.PHONY: lint
lint:
	golangci-lint run ./...

.PHONY: all
all: lint test

.PHONY: test
test:
	go test -race ./...

