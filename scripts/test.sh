#!/bin/bash

REDIS_ENABLED=true incus &
INCUSPID=$!

go test -v ./incustest
RV=$?

if [[ $RV ]]; then 
    kill $INCUSPID
    exit $RV
fi

# for some reason go test -bench only works in current directory?
cd incustest
go test -bench -v .
RV=$?

if [[ $RV ]]; then 
    kill $INCUSPID
    exit $RV
fi

kill $INCUSPID
exit 0
