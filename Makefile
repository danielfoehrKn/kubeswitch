.PHONY: format
format:
	@./hack/format.sh

.PHONY: check
check:
	@./hack/check.sh

.PHONY: build
build: build-switcher build-hooks

.PHONY: build-switcher
build-switcher:
	@go build -o switcher ./cmd/main.go

.PHONY: build-hooks
build-hooks:
	@go build -o hook-gardener-landscape-sync hooks/gardener-landscape-sync/cmd/main.go

.PHONY: all
all: format check build

.PHONY: revendor
revendor:
	@GO111MODULE=on go mod vendor
	@GO111MODULE=on go mod tidy

.PHONY: install-requirements
install-requirements:
	@./hack/install-requirements.sh