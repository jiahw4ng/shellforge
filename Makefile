SANDBOX_IMAGE ?= shellforge-sandbox:0.1.0
BIN_DIR ?= bin
APP_BINARY ?= $(BIN_DIR)/shellforge
WINDOWS_APP_BINARY ?= $(BIN_DIR)/shellforge.exe

.PHONY: build build-windows fmt fmt-check lint lessonlint test check sandbox-image

build:
	mkdir -p $(BIN_DIR)
	go build -trimpath -ldflags="-s -w" -o $(APP_BINARY) ./cmd/shellforge

build-windows:
	mkdir -p $(BIN_DIR)
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -trimpath -ldflags="-s -w" -o $(WINDOWS_APP_BINARY) ./cmd/shellforge

fmt:
	golangci-lint fmt

fmt-check:
	golangci-lint fmt --diff

lint:
	golangci-lint run

lessonlint:
	go run ./cmd/lessonlint

test:
	go test -count=1 ./...

check: fmt-check lint lessonlint test

sandbox-image:
	docker build --file docker/lesson.Dockerfile --platform linux/amd64 --tag $(SANDBOX_IMAGE) .
