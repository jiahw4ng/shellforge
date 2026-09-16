SANDBOX_IMAGE ?= shellforge-sandbox:0.1.0

.PHONY: fmt fmt-check lint lessonlint test check sandbox-image

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
