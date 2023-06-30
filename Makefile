.PHONY: run fmt build all clean
APP_NAME := keeper
PKG_NAME := "khromalabs/keeper"
# SOURCE = $(wildcard *.go)
MSG_DELETED_EXEC='Borrado ejecutable "$(APP_NAME)"'
MSG_DONE_NOTHING='No se hace nada'

all: build
run: fmt
	@KEEPER_ENABLE_DEBUG=1 go run $(PKG_NAME) $(ARG)
fmt:
	@go fmt *.go
tidy:
	@go mod tidy
build: fmt tidy
	@go build -o $(APP_NAME) .
lint:
	@go vet .
clean:
	@[ -f $(APP_NAME) ] && (rm -f $(APP_NAME); echo ${MSG_DELETED_EXEC}) || echo ${MSG_DONE_NOTHING}
