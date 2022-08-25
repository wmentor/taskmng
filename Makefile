.PHONY: all

all: generate gofmt tests linter

linter:
	@if [ -x "$$(command -v golangci-lint)" ]; then echo "Run linter..." ; golangci-lint run ; else echo "golangci-lint not found"; fi

tests:
	@echo "Run test..."
	@go test ./... -cover

gofmt:
	@if [ -x "$$(command -v gofmt)" ]; then echo "Run gofmt..." ; gofmt -w -s . ; else echo "gofmt not found"; fi

generate:
	@if [ -x "$$(command -v mockgen)" ]; then echo "Go generate..." ; go generate ./... ; else echo "mockgen not found"; fi
