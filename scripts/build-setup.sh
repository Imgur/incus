#!/bin/bash

go get -v golang.org/x/tools/cmd/vet
if ! go get -v code.google.com/p/go.tools/cmd/cover; then go get -v golang.org/x/tools/cmd/cover; fi
go get -v github.com/axw/gocov/gocov
go get -v github.com/mattn/goveralls
