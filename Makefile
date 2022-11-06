SHELL := bash

.PHONY: install-deps
install-deps:
	@echo "==> Installing dependencies"
	go install github.com/goreleaser/goreleaser@latest

.PHONY: build
build:
	@echo "==> Build Modelbox"
	goreleaser release --snapshot --rm-dist


.PHONY: sync-docker-hub
sync-docker-hub:
	@echo "==> Sync with docker hub"
	docker push modelboxdotio/modelbox:0.0.1-next-arm64v8
	docker push modelboxdotio/modelbox:0.0.1-next-amd64
	docker manifest create modelboxdotio/modelbox:latest modelboxdotio/modelbox:0.0.1-next-arm64v8 modelboxdotio/modelbox:0.0.1-next-amd64
	docker manifest push --purge modelboxdotio/modelbox:latest

.PHONY: test
test-server:
	@echo "==> Test Modelbox Server"
	go test ./server/storage/...

.PHONY: install-sdk-py
install-sdk-py:
	@echo "==> Installing Python SDK"
	cd sdk-py && pip install .

.PHONY: test-sdk-py
test-sdk-py:
	@echo "==> Testing python sdk"
	cd sdk-py && pip install .
	cd sdk-py && python tests/test_modelbox_api.py

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
upload-sdk-py:
	@echo "===> Uploading to pypi"
	pip install twine
	cd sdk-py && twine upload dist/*


.PHONY: update-protos
update-protos:
	@echo "==> Updating protos"
	./proto/gen_proto.sh

.PHONY: gen-static
gen-static:
	@echo "==> Generating static files" 
	cd cmd/modelbox && go-bindata assets/...
