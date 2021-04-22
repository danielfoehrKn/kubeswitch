.PHONY: format
format:
	@./hack/format.sh ./cmd ./pkg

.PHONY: test
test:
	@./hack/test.sh ./cmd/... ./pkg/...

.PHONY: check
check:
	@./hack/test.sh ./cmd/... ./pkg/...
	@./hack/check.sh ./cmd/... ./pkg/...

.PHONY: build
build: build-switcher build-hooks

.PHONY: build-switcher
build-switcher:
	@env GOOS=linux GOARCH=amd64 go build -ldflags "-w -X github.com/danielfoehrkn/kubeswitch/cmd/switcher.version=${APP_VERSION} -X github.com/danielfoehrkn/kubeswitch/cmd/switcher.buildDate=${BUILD_TIME}" -o hack/switch/switcher_linux_amd64 ./cmd/main.go
	@env GOOS=darwin GOARCH=amd64 go build -ldflags "-w -X github.com/danielfoehrkn/kubeswitch/cmd/switcher.version=${APP_VERSION} -X github.com/danielfoehrkn/kubeswitch/cmd/switcher.buildDate=${BUILD_TIME}" -o hack/switch/switcher_darwin_amd64 ./cmd/main.go

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