#!/usr/bin/env bash

set -e
set -x
MODE=atomic
echo "mode: $MODE" >coverage.txt

# Packages that have any tests.
PKG=$(go list -f '{{if .TestGoFiles}} {{.ImportPath}} {{end}}' ./...)

go test $PKG

for d in $PKG; do
	go test -race -coverprofile=profile.out -covermode=$MODE $d
	if [ -f profile.out ]; then
		cat profile.out | grep -v "^mode: " >>coverage.txt
		rm profile.out
	fi
done

go vet -all ./...
if [ "$RUN_GOLANGCI_LINTER" != "false" ]; then
	golangci-lint run -D errcheck -exlude=errcheck_excludes.txt --timeout=1m ./...
fi

gofmt -s -d .
