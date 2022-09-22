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

.PHONY: build-sdk-py
build-sdk-py:
                @echo "==> Building modelbox py distribution "
				cd sdk-py && python -m build .

.PHONY: upload-sdk-py-test
upload-sdk-py-test:
                @echo "===> Uploading to test.pypi"
				pip install twine
				cd sdk-py && twine upload --repository testpypi dist/*

.PHONY: upload-sdk-py
upload-sdk-py-test:
                @echo "===> Uploading to pypi"
				pip install twine
				cd sdk-py && twine upload --repository dist/*