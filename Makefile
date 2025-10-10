.PHONY: build test
build:
	go build ./cmd/dragon
test:
	go test ./...
