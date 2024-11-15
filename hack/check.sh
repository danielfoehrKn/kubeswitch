#!/bin/bash
set -e

echo "> Check"

echo "Executing golangci-lint"
which golangci-lint
golangci-lint run "${SOURCE_TREES[@]}" --timeout=10m0s --verbose --print-resources-usage --modules-download-mode=vendor

echo "Executing go vet"
go vet -mod=vendor $@

echo "Executing gofmt/goimports"
folders=()
for f in $@; do
  folders+=( "$(echo $f | sed 's/\.\/\(.*\)\/\.\.\./\1/')" )
done
unformatted_files="$(goimports -l ${folders[*]})"
if [[ "$unformatted_files" ]]; then
  echo "Unformatted files detected:"
  echo "$unformatted_files"
  exit 1
fi

echo "Check for license headers"
addlicense -check -ignore ".git/**" -ignore "vendor/**" -ignore "hack/**" -ignore "**/*.yaml" -ignore "**/*.yml" -ignore "resources/demo-config-files/**" -ignore "**/*.proto" .

echo "All checks successful"
