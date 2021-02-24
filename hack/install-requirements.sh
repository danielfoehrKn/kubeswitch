#!/bin/bash
set -e

DIRNAME="$(echo "$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )")"

cd "$DIRNAME/.."
export GO111MODULE=on
echo "Installing requirements"

go get -u github.com/danielfoehrKn/importsort@befc7d7538f4702dbaed7951bc60577e7f567237
curl -sfL "https://install.goreleaser.com/github.com/golangci/golangci-lint.sh" | sh -s -- -b $(go env GOPATH)/bin v1.18.0

# license header
go get -u github.com/google/addlicense