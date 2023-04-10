.PHONY: run fmt build all clean
APP_NAME := keeper
# SOURCE = $(wildcard *.go)
MSG_DELETED_EXEC='Borrado ejecutable "$(APP_NAME)"'
MSG_DONE_NOTHING='No se hace nada'
MSG_DONE_EXEC="Aplicación \"$(APP_NAME)\" compilada con éxito"

run:
	@go run -o $(APP_NAME) . $(ARG)
	$(APP_NAME) $(ARG)
fmt:
	@go fmt *.go
build: fmt
	@go build -o $(APP_NAME) .
# @echo "${MSG_DONE_EXEC}"
all: build
lint:
	@go vet .
clean:
	@[ -f $(APP_NAME) ] && (rm -f $(APP_NAME); echo ${MSG_DELETED_EXEC}) || echo ${MSG_DONE_NOTHING}
