.PHONY: usage build lint lint-autofix test test-coverage test-show-coverage run tidy release release-all

OK_COLOR=\033[32;01m
NO_COLOR=\033[0m
ERROR_COLOR=\033[31;01m
WARN_COLOR=\033[33;01m

PKGS = $(shell go list ./...)
BIN = git-sv

ECHOFLAGS ?=

BUILD_TIME = $(shell date +"%Y%m%d%H%M")
VERSION ?= dev-$(BUILD_TIME)

BUILDOS ?= linux
BUILDARCH ?= amd64
BUILDENVS ?= CGO_ENABLED=0 GOOS=$(BUILDOS) GOARCH=$(BUILDARCH)
BUILDFLAGS ?= -a -installsuffix cgo --ldflags '-X main.Version=$(VERSION) -extldflags "-lm -lstdc++ -static"'

COMPRESS_TYPE ?= targz

usage: Makefile
	@echo $(ECHOFLAGS) "to use make call:"
	@echo $(ECHOFLAGS) "    make <action>"
	@echo $(ECHOFLAGS) ""
	@echo $(ECHOFLAGS) "list of available actions:"
	@sed -n 's/^##//p' $< | column -t -s ':' | sed -e 's/^/ /'

## build: build git-sv
build: test
	@echo $(ECHOFLAGS) "$(OK_COLOR)==> Building binary ($(BUILDOS)/$(BUILDARCH)/$(BIN))...$(NO_COLOR)"
	@$(BUILDENVS) go build -v $(BUILDFLAGS) -o bin/$(BUILDOS)_$(BUILDARCH)/$(BIN) ./cmd/git-sv

## lint: run golangci-lint without autofix
lint:
	@echo $(ECHOFLAGS) "$(OK_COLOR)==> Running golangci-lint...$(NO_COLOR)"
	@golangci-lint run ./... --config .golangci.yml

## lint-autofix: run golangci-lint with autofix enabled
lint-autofix:
	@echo $(ECHOFLAGS) "$(OK_COLOR)==> Running golangci-lint...$(NO_COLOR)"
	@golangci-lint run ./... --config .golangci.yml --fix

## test: run unit tests
test:
	@echo $(ECHOFLAGS) "$(OK_COLOR)==> Running tests...$(NO_COLOR)"
	@go test $(PKGS)

## test-coverage: run tests with coverage
test-coverage:
	@echo $(ECHOFLAGS) "$(OK_COLOR)==> Running tests with coverage...$(NO_COLOR)"
	@go test -race -covermode=atomic -coverprofile coverage.out ./...

## test-show-coverage: show coverage
test-show-coverage: test-coverage
	@echo $(ECHOFLAGS) "$(OK_COLOR)==> Show test coverage...$(NO_COLOR)"
	@go tool cover -html coverage.out

## run: run git-sv
run:
	@echo $(ECHOFLAGS) "$(OK_COLOR)==> Running bin/$(BUILDOS)_$(BUILDARCH)/$(BIN)...$(NO_COLOR)"
	@./bin/$(BUILDOS)_$(BUILDARCH)/$(BIN) $(args)

## tidy: execute go mod tidy
tidy:
	@echo $(ECHOFLAGS) "$(OK_COLOR)==> runing tidy"
	@go mod tidy

## release: prepare binary for release
release:
	make build
ifeq ($(COMPRESS_TYPE), zip)
	@zip -j bin/git-sv_$(VERSION)_$(BUILDOS)_$(BUILDARCH).zip bin/$(BUILDOS)_$(BUILDARCH)/$(BIN)
else
	@tar -czf bin/git-sv_$(VERSION)_$(BUILDOS)_$(BUILDARCH).tar.gz -C bin/$(BUILDOS)_$(BUILDARCH)/ $(BIN)
endif

## release-all: prepare linux, darwin and windows binary for release (requires sv4git)
release-all:
	@rm -rf bin

	VERSION=$(shell git sv nv)                   BUILDOS=linux   BUILDARCH=amd64 make release
	VERSION=$(shell git sv nv)                   BUILDOS=darwin  BUILDARCH=amd64 make release
	VERSION=$(shell git sv nv) COMPRESS_TYPE=zip BUILDOS=windows BUILDARCH=amd64 make release

	VERSION=$(shell git sv nv)                   BUILDOS=linux   BUILDARCH=arm64 make release
	VERSION=$(shell git sv nv)                   BUILDOS=darwin  BUILDARCH=arm64 make release
	VERSION=$(shell git sv nv) COMPRESS_TYPE=zip BUILDOS=windows BUILDARCH=arm64 make release
