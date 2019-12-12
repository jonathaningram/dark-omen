GO ?= go

test:
	$(GO) test -race ./...

.PHONY: test
