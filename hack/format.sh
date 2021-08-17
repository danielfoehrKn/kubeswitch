#!/bin/bash
set -e

echo "> Format"

goimports -l -w $@

addlicense -c "The Kubeswitch authors" pkg/
addlicense -c "The Kubeswitch authors" cmd/
addlicense -c "The Kubeswitch authors" hooks/
addlicense -c "The Kubeswitch authors" types/
