SHELL := bash

.PHONY: install-deps
install-deps:
				@echo "==> Installing dependencies"
				go install github.com/goreleaser/goreleaser@latest

.PHONY: build
build:
		@echo "==> Build Modelbox"
		goreleaser release --snapshot --rm-dist

.PHONY: test
test-server:
			  @echo "==> Test Modelbox Server"
			  go test ./server/storage/...

.PHONY: install-sdk-py 
install-sdk-py:
				@echo "==> Installing Python SDK"
				cd sdk-py && pip install .