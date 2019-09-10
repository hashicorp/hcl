#!/bin/bash

set -e
echo "" > coverage.txt

for d in $(go list ./... | grep -v vendor); do
    go test -coverprofile=/artifacts/profile.out -covermode=atomic $d
    if [ -f /artifacts/profile.out ]; then
        cat /artifacts/profile.out >> /artifacts/coverage.txt
        rm /artifacts/profile.out
    fi
done
