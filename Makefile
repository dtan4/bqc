NAME := bqc

GO_BUILD_ENVS := CGO_ENABLED=1 CXX=clang++

.DEFAULT_GOAL := build

.PHONY: build
build:
	$(GO_BUILD_ENVS) go build -o bin/$(NAME)

.PHONY: clean
clean:
	rm -rf bin/*

.PHONY: install
install:
	$(GO_BUILD_ENVS) go install

.PHONY: test
test:
	$(GO_BUILD_ENVS) go test -race ./...
