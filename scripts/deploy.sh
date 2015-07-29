#!/bin/bash

VERSION=$(date +%s)

aws s3 cp $GOPATH/incus.tar s3://imgur-incus/incus-latest.tar && \
    aws s3 cp $GOPATH/incus.tar s3://imgur-incus/incus-$VERSION.tar && \
    aws deploy create-deployment --application-name Incus --deployment-group-name Production --s3-location bundleType=tar,bucket=imgur-incus,key=incus-$VERSION.tar && \
