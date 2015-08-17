#!/bin/bash

set -e

start_incus() {
    REDIS_ENABLED=true CONNECTION_TIMEOUT=10000 incus &
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
