.PHONY: format
format:
	@./hack/format.sh

.PHONY: check
check:
	@./hack/check.sh

.PHONY: build
build:
	@go build

.PHONY: all
all: format check build

.PHONY: revendor
revendor:
	@GO111MODULE=on go mod vendor
	@GO111MODULE=on go mod tidy

.PHONY: install-requirements
install-requirements:
	@./hack/install-requirements.sh