NAME := bqc

.DEFAULT_GOAL := build

.PHONY: build
build:
	go build -o bin/$(NAME)

.PHONY: clean
clean:
	rm -rf bin/*

.PHONY: install
install:
	go install -ldflags="-s -w"

.PHONY: test
test:
	go test -race ./...
