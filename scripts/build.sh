#!/bin/bash

set -v

# Exclude ./incustest from the go test run because it requires a redis server
BUILDR=". ./incus"

rm -f $GOPATH/bin/{incus,incustest} $GOPATH/incus.tar && \
    go get -d -v ./... && \
    go install -ldflags "-X main.builddate `date -u +.%Y%m%d.%H%M%S`" $BUILDR && \
    go vet ./...  && \
    go test -v $BUILDR && \
    tar cf $GOPATH/incus.tar -C $GOPATH bin/incus -C $GOPATH/src/github.com/Imgur/incus scripts/ appspec.yml

exit 0
