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