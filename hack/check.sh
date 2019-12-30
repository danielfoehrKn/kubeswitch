#!/bin/bash
set -e

DIRNAME="$(echo "$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )")"

echo "Executing golangci-lint"
golangci-lint run "${SOURCE_TREES[@]}"

echo "Checking for format issues with importsort"
unsorted_files="$(importsort -l ./main.go)"
if [[ "$unsorted_files" ]]; then
    echo "Unformatted files detected:"
    echo "$unsorted_files"
    exit 1
fi
echo "All checks successful"
