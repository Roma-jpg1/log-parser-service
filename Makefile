GO ?= go
GOFMT ?= gofmt

.PHONY: fmt fmt-check lint test check

fmt:
	$(GOFMT) -w .

fmt-check:
	@test -z "$$($(GOFMT) -l .)" || (echo "gofmt is required for:"; $(GOFMT) -l .; exit 1)

lint:
	$(GO) vet ./...

test:
	$(GO) test ./...

check: fmt-check lint test
