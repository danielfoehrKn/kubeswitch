#!/bin/bash

set -e

echo "> Test"

GO111MODULE=on go test -race -mod=vendor $@ | grep -v 'no test files'
