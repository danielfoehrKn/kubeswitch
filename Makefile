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
	@env GOOS=linux GOARCH=amd64 go build -o hack/switch/switcher_linux_amd64 ./cmd/main.go
	@env GOOS=darwin GOARCH=amd64 go build -o hack/switch/switcher_darwin_amd64 ./cmd/main.go

.PHONY: build-hooks
build-hooks:
	@env GOOS=linux GOARCH=amd64 go build -o hack/hooks/hook_gardener_landscape_sync_linux_amd64 hooks/gardener-landscape-sync/cmd/main.go
	@env GOOS=darwin GOARCH=amd64 go build -o hack/hooks/hook_gardener_landscape_sync_darwin_amd64 hooks/gardener-landscape-sync/cmd/main.go

.PHONY: all
all: format check build

.PHONY: revendor
revendor:
	@GO111MODULE=on go mod vendor
	@GO111MODULE=on go mod tidy

.PHONY: install-requirements
install-requirements:
	@./hack/install-requirements.sh