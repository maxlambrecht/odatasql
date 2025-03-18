GO := go

.PHONY: test
test:
	$(GO) test ./...

.PHONY: deps
deps:
	$(GO) mod tidy
