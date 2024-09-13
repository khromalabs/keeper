.PHONY: run fmt build all clean
APP_NAME := keeper
PKG_NAME := "khromalabs/keeper"
# SOURCE = $(wildcard *.go)
MSG_DELETED_EXEC='Deleted executable "$(APP_NAME)"'
MSG_DONE_NOTHING='Nothing to do'
PREFIX ?= /usr
XDG_DATA_HOME ?= ~/.local/share

all: build
run: fmt
	@KEEPER_ENABLE_DEBUG=1 go run $(PKG_NAME) $(ARG)
fmt:
	@go fmt *.go
tidy:
	@go mod tidy
build: fmt tidy
	@go build -o $(APP_NAME) -ldflags "-X main.version=$(shell git describe --tags --dirty --always)" .
lint:
	@go vet .
clean:
	@[ -f $(APP_NAME) ] && (rm -f $(APP_NAME); echo ${MSG_DELETED_EXEC}) || echo ${MSG_DONE_NOTHING}

install: build
	@install -Dm 755 $(APP_NAME) $(PREFIX)/bin/$(APP_NAME)
	@install -Dm 644 resources/completion/keeper $(PREFIX)/share/bash-completion/completions/keeper
	@install -Dm 644 resources/man/keeper.1 $(PREFIX)/share/man/man1/keeper.1
	@mkdir -p $(PREFIX)/share/$(APP_NAME)
	@install -Dm 644 resources/templates/* $(XDG_DATA_HOME)/$(APP_NAME)/templates

version:
	@echo $(shell git describe --tags --dirty --always)
