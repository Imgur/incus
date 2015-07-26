Incus [![Build Status](https://travis-ci.org/Imgur/incus.svg?branch=master)](https://travis-ci.org/Imgur/incus) [![Coverage Status](https://coveralls.io/repos/Imgur/incus/badge.svg)](https://coveralls.io/r/Imgur/incus)
=========

![incus](http://i.imgur.com/7ZgRrA5.png)

middleware for distributing messages via websockets, long polling, and push notifications

## Features
* Websocket authentication and management
* Long Poll fall back
* iOS push notifications support through APNS
* Android push notifications support through GCM
* Routing messages to spefific phone, authenticated user, or webpage url
* Configurable option for allowing users to send messages to other users
* Redis pub/sub and Redis List support for sending messages from an application
* SSL support
* Stats logging

## Usage

### incus.js
Once incus is running you can include incus.js onto your site.
Here's the basic usage of incus.js

```Javascript
var socket = new Incus('http://localhost:4000', 'UID', '/page/path');
socket.on('connect', function() {
    console.log('connected');
}

socket.on('Event', function(data) {
    alert(data);
}

socket.on('Event1', function (data) {
   console.log(data);
}

socket.on('Event2', function(data) {
    console.log('neat');
} 

var data = {data: 'dummy-data'};

$('#button').on('click', function() {
    socket.MessageUser('Event', 'UID', data); 
    socket.MessageAll('Event1', data);
    socket.MessagePage('Event2', '/page/path', data);
});
```

### Application to a web browser

UID is a unique identifier for the user. It can be anywhere from an auto incremented ID to something more private such as a session id or OAuth token.

To send events to Incus from your webapp you need to publish a json formated string to a **Redis pub/sub channel** that Incus is listening on. This channel key can be donfigured but defaults to `Incus`. The json format is as follows:

```Javascript
{
    'command' : {
        'command' : string (message|setpage),
        'user'    : (optional) string -- UID,
        'page'    : (optional) string -- page identifier
    },
    'message' : {
        'event' : string,
        'data'  : object,
        'time'  : int
    }
}
```

the command is used to route the message to the correct user.
* if user and page are both unset the message object will be sent to all users
* if both user and page are set the message object will be sent to that user on that page
* if just user is set, the message object will be sent to all sockets that user has connected
* if just page is set, the message object will be sent to all sockets whose page matches the page identifier


### Push notifications 
To send push notifications from your app, you need to push a json formated string to a **Redis list**. The list key is configurable but defaults to `Incus_Queue`

Android and iOS have slightly differnt schemas for sending push notifications.

#### iOS:
```Javascript
{
    'command' : {
        'command'      : 'pushios',
        'device_token' : string -- device token registered with APNS ,
        'build'        : string -- build environment (store|beta|enterprise|development)
    },
    'message' : {
        'event' : string,
        'data'  : {
            'badge_count': int,
            ...,
            ...,
        },
        'time'  : int
    }
}
```

#### Android:

Multiple registration ids can be listed in the same command

```Javascript
{
    'command' : {
        'command'          : 'pushandroid',
        'registration_ids' : string -- one or more registration ids separated by commas
    },
    'message' : {
        'event' : string,
        'data'  : object,
        'time'  : int
    }
}
```

#### APNS and GCM errors

You should follow APNS' guidlines on failed push attempts, they require querying their feedback service daily to find bad device tokens. See for more details: https://developer.apple.com/library/ios/documentation/NetworkingInternet/Conceptual/RemoteNotificationsPG/Chapters/CommunicatingWIthAPS.html

The GCM service doesn't offer a feedback service, so when a push fails incus will add all relevant information to an error list in Redis (defaults to `Incus_Android_Error_Queue`). This should be used to remove bad registration ids from your app. 

## Installation
### Method 1: Docker
Install [Docker](https://docs.docker.com/installation/#installation)

Start an instance of Redis:

```Shell
docker run -d --name incusredis redis
```

Start Incus:

```Shell
docker run -d --link incusredis:redis --name incus \
        -p 4000:4000 jwgur/incus 
```

### Method 2: Source
Install GO: http://golang.org/doc/install

Clone the repo:
```Shell
mkdir $GOPATH/src/github.com/Imgur/
git clone git@github.com:Imgur/incus.git
```

To Install:
```Shell
go get -v ./...
go install -v ./incus
cp $GOPATH/bin/incus /usr/sbin/incus
cp scripts/initd.sh /etc/init.d/incus
mkdir /etc/incus
cp incus.conf /etc/incus/incus.conf
touch /var/log/incus.log
```
