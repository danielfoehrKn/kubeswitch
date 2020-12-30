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
	@go build -o hack/switch/switcher ./cmd/main.go
	@env GOOS=linux GOARCH=amd64 go build -o hack/switch/switcher_linux_amd64 ./cmd/main.go

.PHONY: build-hooks
build-hooks:
	@go build -o hack/hooks/hook_gardener_landscape_sync hooks/gardener-landscape-sync/cmd/main.go
	@env GOOS=linux GOARCH=amd64 go build -o hack/hooks/hook_gardener_landscape_sync_linux_amd64 hooks/gardener-landscape-sync/cmd/main.go

.PHONY: all
all: format check build

.PHONY: revendor
revendor:
	@GO111MODULE=on go mod vendor
	@GO111MODULE=on go mod tidy

.PHONY: install-requirements
install-requirements:
	@./hack/install-requirements.sh