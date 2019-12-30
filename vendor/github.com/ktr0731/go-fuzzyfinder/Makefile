SHELL := /bin/bash

export PATH := $(PWD)/_tools:$(PATH)
export GO111MODULE := on

.PHONY: generate
generate:
	go generate ./...

.PHONY: dept
dept:
	@go get github.com/ktr0731/dept@v0.1.2
	@go build -o _tools/dept github.com/ktr0731/dept

.PHONY: tools
tools: dept
	@dept -v build

.PHONY: build
build:
	go build ./...

.PHONY: test
test: format unit-test credits

.PHONY: format
format:
	go mod tidy

.PHONY: credits
credits:
	gocredits . > CREDITS

.PHONY: unit-test
unit-test: lint
	go test -v -race ./...

.PHONY: lint
lint:
	golangci-lint run --disable-all \
		--skip-files 'helper_test.go' \
		-e 'should have name of the form ErrFoo' -E 'deadcode,govet,golint' \
		./...

.PHONY: coverage
coverage:
	DEBUG=true go test -v -coverpkg ./... -covermode=atomic -coverprofile=coverage.txt -race $(shell go list ./...)

.PHONY: coverage-web
coverage-web: coverage
	go tool cover -html=coverage.txt
