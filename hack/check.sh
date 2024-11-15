#!/bin/bash
set -e

echo "> Check"

echo "Executing golangci-lint"
which golangci-lint
golangci-lint run "${SOURCE_TREES[@]}" --timeout=10m0s --verbose --print-resources-usage --modules-download-mode=vendor

echo "Check for license headers"
addlicense -check -ignore ".git/**" -ignore "vendor/**" -ignore "hack/**" -ignore "**/*.yaml" -ignore "**/*.yml" -ignore "resources/demo-config-files/**" -ignore "**/*.proto" .

echo "All checks successful"
