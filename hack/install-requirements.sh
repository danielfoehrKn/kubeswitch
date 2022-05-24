#!/bin/bash
set -e

DIRNAME="$(echo "$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )")"

cd "$DIRNAME/.."
export GO111MODULE=on
echo "Installing requirements"

curl -sfL "https://install.goreleaser.com/github.com/golangci/golangci-lint.sh" | sh -s -- -b $(go env GOPATH)/bin v1.42.0

go get golang.org/x/tools/cmd/goimports

# license header
# go get -u github.com/google/addlicense