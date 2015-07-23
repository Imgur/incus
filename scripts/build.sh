#!/bin/bash

go get -v golang.org/x/tools/cmd/vet
if ! go get -v code.google.com/p/go.tools/cmd/cover; then go get -v golang.org/x/tools/cmd/cover; fi
go get -v github.com/axw/gocov/gocov
go get -v github.com/mattn/goveralls

# Exclude ./incustest from the go test run because it requires a redis server
BUILDR=". ./incus"

go get -v ./...
go install -v $BUILDR && go vet ./... && go test -v $BUILDR && tar cf $GOPATH/incus.tar -C $GOPATH bin/incus -C $GOPATH/src/github.com/Imgur/incus scripts/ appspec.yml
