#!/bin/bash

# Exclude ./incustest from the go test run because it requires a redis server
BUILDR=". ./incus"

go get -v ./...
go install -v $BUILDR && go vet ./... && go test -v $BUILDR
