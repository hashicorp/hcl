TEST?=./...

default: test

fmt: generate
	go fmt ./...

test: generate
	go test $(TEST) $(TESTARGS)

generate:
	go generate ./...

.PHONY: default generate test
