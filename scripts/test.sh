#!/bin/bash
REDIS_ENABLED=true incus &
sleep 1
go test -v ./incustest
RV=$?
INCUSPID=$!
kill $INCUSPID
exit $RV
