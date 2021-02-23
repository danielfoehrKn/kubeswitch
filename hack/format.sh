#!/bin/bash
set -e

echo "> Format"

goimports -l -w $@
