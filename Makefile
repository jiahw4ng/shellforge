.PHONY: fmt fmt-check lint test check

fmt:
	golangci-lint fmt

fmt-check:
	golangci-lint fmt --diff

lint:
	golangci-lint run ./...

test:
	go test ./...

check: fmt-check lint test
