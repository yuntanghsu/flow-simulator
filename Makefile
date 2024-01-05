GO                 ?= go
BINDIR := $(CURDIR)/bin

.PHONY: tidy
tidy:
	go mod tidy

.PHONY: bin
bin:
	@GOOS=$* $(GO) build -o bin/main ./cmd/main.go || (echo an error while building binary, exiting!; sh -c 'exit 1';)
