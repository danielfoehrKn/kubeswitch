#!/bin/bash
set -e

go fmt ./...

importsort -w ./pkg
importsort -w ./cmd
