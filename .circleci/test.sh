#!/bin/bash

set -e
echo "" > coverage.txt

for d in $(go list ./... | grep -v vendor); do
    go test -coverprofile=/tmp/artifacts/profile.out -covermode=atomic $d
    if [ -f /tmp/artifacts/profile.out ]; then
        cat /tmp/artifacts/profile.out >> /tmp/artifacts/coverage.txt
        rm /tmp/artifacts/profile.out
    fi
done
