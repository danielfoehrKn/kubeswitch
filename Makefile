DATE=$(shell date -u +%Y-%m-%d)
VERSION=$(shell cat VERSION | sed 's/-dev//g')

#########################################
# Tools                                 #
#########################################

TOOLS_DIR := hack/tools
include hack/tools.mk

#########################################
# Targets                                 #
#########################################

.PHONY: format
format: $(GOLICENSES) $(GOIMPORTS)
	@./hack/format.sh ./cmd ./pkg

.PHONY: test
test:
	@./hack/test.sh ./pkg/...

.PHONY: check
check: $(GOIMPORTS) $(GOLANGCI_LINT) $(ADDLICENSE)
	@./hack/test.sh ./pkg/...
	@./hack/check.sh ./cmd/... ./pkg/...

.PHONY: build
build: build-switcher

.PHONY: build-switcher
build-switcher:
	@env GOOS=linux GOARCH=amd64 go build -ldflags "-w -X github.com/danielfoehrkn/kubeswitch/cmd/switcher.version=${VERSION} -X github.com/danielfoehrkn/kubeswitch/cmd/switcher.buildDate=${DATE}" -o hack/switch/switcher_linux_amd64 ./cmd/main.go
	@env GOOS=linux GOARCH=arm64 go build -ldflags "-w -X github.com/danielfoehrkn/kubeswitch/cmd/switcher.version=${VERSION} -X github.com/danielfoehrkn/kubeswitch/cmd/switcher.buildDate=${DATE}" -o hack/switch/switcher_linux_arm64 ./cmd/main.go
	@env GOOS=darwin GOARCH=amd64 go build -ldflags "-w -X github.com/danielfoehrkn/kubeswitch/cmd/switcher.version=${VERSION} -X github.com/danielfoehrkn/kubeswitch/cmd/switcher.buildDate=${DATE}" -o hack/switch/switcher_darwin_amd64 ./cmd/main.go
	@env GOOS=darwin GOARCH=arm64 go build -ldflags "-w -X github.com/danielfoehrkn/kubeswitch/cmd/switcher.version=${VERSION} -X github.com/danielfoehrkn/kubeswitch/cmd/switcher.buildDate=${DATE}" -o hack/switch/switcher_darwin_arm64 ./cmd/main.go

.PHONY: all
all: format check build

.PHONY: revendor
revendor:
	@GO111MODULE=on go mod vendor
	@GO111MODULE=on go mod tidy