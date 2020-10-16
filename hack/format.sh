#!/bin/bash
set -e

importsort -w ./cmd/main.go
importsort -w ./cmd/switcher/switch.go
importsort -w ./pkg/clean.go
importsort -w ./pkg/main.go
importsort -w ./pkg/hooks.go
