#!/bin/bash

VERSION=$(date +%s)

if [ "${TRAVIS_PULL_REQUEST}" = "false" ] && [ "${TRAVIS_BRANCH}" = "master" ]; then
    curl -X POST -H "Content-type: application/json" \
            -d '{
        "title": "Incus deployed",
        "text": "A new version of Incus was deployed to AWS CodeDeploy",
        "priority": "normal",
        "tags": ["env:prod"],
        "alert_type": "info"
    }' \
        "https://app.datadoghq.com/api/v1/events?api_key=$DATADOG_API_KEY"

    aws s3 cp $GOPATH/incus.tar s3://imgur-incus/incus-latest.tar && \
        aws s3 cp $GOPATH/incus.tar s3://imgur-incus/incus-$VERSION.tar && \
        aws deploy create-deployment --application-name Incus --deployment-group-name Production --s3-location bundleType=tar,bucket=imgur-incus,key=incus-$VERSION.tar
fi
