.PHONY: fmt build test snapshot clean

APP := imprun
CMD := ./cmd/imprun
VERSION ?= dev
BIN_DIR ?= .tmp/bin
ifeq ($(OS),Windows_NT)
EXEEXT := .exe
endif
BIN := $(BIN_DIR)/$(APP)$(EXEEXT)

fmt:
	go fmt ./...

build:
	@mkdir -p "$(BIN_DIR)"
	go build -trimpath -ldflags "-X github.com/imprun/cli/internal/controlcli.Version=$(VERSION)" -o "$(BIN)" $(CMD)

test:
	go test ./...

snapshot:
	goreleaser release --snapshot --clean

clean:
	rm -rf .tmp dist
