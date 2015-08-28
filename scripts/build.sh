#!/bin/bash

set -v

# Exclude ./incustest from the go test run because it requires a redis server
BUILDR=". ./incus"

BUILDDATE="$(date -u +.%Y%m%d.%H%M%S)"
BUILDVAR="$BUILDDATE-$(git rev-parse HEAD)"

if [ "0" != "$(git status --porcelain | wc -l)" ]; then 
    # The build version will have a star at the end if it was built from an unclean directory.
    BUILDVAR="$BUILDVAR*"
fi

rm -f $GOPATH/bin/{incus,incustest,config.yml} $GOPATH/incus.tar && \
	cp $GOPATH/src/github.com/Imgur/incus/scripts/test_config/config.yml $GOPATH/bin/

    go get -d -v ./... && \
    go install -ldflags "-X main.BUILD $BUILDVAR" $BUILDR && \
    go vet ./...  && \
    go test -v $BUILDR && \
    tar cf $GOPATH/incus.tar -C $GOPATH bin/incus -C $GOPATH/src/github.com/Imgur/incus scripts/ appspec.yml

exit 0
