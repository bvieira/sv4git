.PHONY: usage build test run tidy release release-all

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

## test: run unit tests
test:
	@echo $(ECHOFLAGS) "$(OK_COLOR)==> Running tests...$(NO_COLOR)"
	@go test $(PKGS)

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
	@zip -j bin/git-sv_$(VERSION)_$(BUILDOS)_$(BUILDARCH).zip bin/$(BUILDOS)_$(BUILDARCH)/$(BIN)

## release-all: prepare linux, darwin and windows binary for release
release-all:
	@rm -rf bin
	BUILDOS=linux make release
	BUILDOS=darwin make release
	BUILDOS=windows make release
