#!/bin/bash
set -e

echo "> Format"

goimports -l -w $@


# please add yourself to the files you have contributed to
addlicense -c "Daniel Foehr" pkg/
addlicense -c "Daniel Foehr" cmd/
addlicense -c "Daniel Foehr" hooks/
addlicense -c "Daniel Foehr" types/
