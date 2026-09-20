SANDBOX_IMAGE ?= shellforge-sandbox:0.1.0
BIN_DIR ?= bin
APP_BINARY ?= $(BIN_DIR)/shellforge

.PHONY: build fmt fmt-check lint lessonlint test check sandbox-image

build:
	mkdir -p $(BIN_DIR)
	go build -o $(APP_BINARY) ./cmd/shellforge

fmt:
	golangci-lint fmt

fmt-check:
	golangci-lint fmt --diff

lint:
	golangci-lint run ./...

lessonlint:
	go run ./cmd/lessonlint

test:
	go test ./...

check: fmt-check lint lessonlint test

sandbox-image:
	docker build --file docker/lesson.Dockerfile --platform linux/amd64 --tag $(SANDBOX_IMAGE) .
