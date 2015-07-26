Incus [![Build Status](https://travis-ci.org/Imgur/incus.svg?branch=master)](https://travis-ci.org/Imgur/incus) [![Coverage Status](https://coveralls.io/repos/Imgur/incus/badge.svg)](https://coveralls.io/r/Imgur/incus)
=========

![incus](http://i.imgur.com/7ZgRrA5.png)

middleware for distributing messages via websockets, long polling, and push notifications

## Features
* Websocket authentication and management
* Long Poll fall back
* iOS push notifications support through APNS
* Android push notifications support through GCM
* Routing messages to specific phone, authenticated user, or webpage url
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

To send events to Incus from your webapp you need to publish a json formated string to a **Redis pub/sub channel** that Incus is listening on. This channel key can be configured but defaults to `Incus`. The json format is as follows:

```Javascript
{
    'command' : {
        'command' : string (message|setpage),
        'user'    : (optional) string -- Unique User ID,
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
* if just user is set, the message object will be sent to all sockets owned by the user identified by UID
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

The GCM service doesn't offer a feedback service. When a push fails, incus will add all relevant information to an error list in Redis (defaults to `Incus_Android_Error_Queue`). This should be used to remove bad registration ids from your app. 

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
cd $GOPATH/src/github.com/Imgur/
git clone git@github.com:Imgur/incus.git
```

To Install:
```Shell
cd $GOPATH/src/github.com/Imgur/incus
go get -v ./...
go install -v ./incus
cp $GOPATH/bin/incus /usr/sbin/incus
cp scripts/initd.sh /etc/init.d/incus
mkdir /etc/incus
cp incus.conf /etc/incus/incus.conf
touch /var/log/incus.log
```

Starting, Stopping, Restarting incus:
```Shell
sudo /etc/init.d/incus start
sudo /etc/init.d/incus stop
sudo /etc/init.d/incus restart
```
## Configuration
You can configure Incus by passing environment variables. Incus needs to be restarted after any variable change.

#### CLIENT_BROADCASTS

**true**
> Clients may send messages to other clients

**false**
> Clients may not send messages to other clients.

Default: true

_________
#### LISTENING_PORT

This value controls the port that Incus binds to (TCP).

Default: 4000

_________
#### CONNECTION_TIMEOUT (unstable)

This value controls how long TCP connections are held open for.

**0**
> Connections are held open forever.

**anything else**
> Connections are held open for this many seconds.

Default: 0

_________
#### LOG_LEVEL

**debug**
> All messages, including errors and debug are printed to standard output.

**error**
> Only errors are printed to standard output.

Default: debug

_________
#### REDIS_PORT_6379_TCP_ADDR

This value controls the TCP address (or hostname) to connect to Redis.

Note: The environment variable name is always REDIS_PORT_6379_TCP_ADDR even if the port is not 6379.

Default: 127.0.0.1

_________
#### REDIS_PORT_6379_TCP_PORT

This value controls the TCP port to connect to Redis.

Note: The environment variable name is always REDIS_PORT_6379_TCP_PORT even if the port is not 6379.

Default: 6379

_________
#### REDIS_MESSAGE_CHANNEL

This value controls the Redis PubSub channel to use.

Default: Incus

_________
#### TLS_ENABLED

This value controls whether the server will also listen on a TLS-enabled port.

**false**
> TLS is disabled, so the server will only listen over its insecure port.

**true**
> TLS is enabled, so the socket will listen over both its insecure port and its TLS-enabled secure port.

Default: false

_________
#### TLS_PORT

This value controls what TCP port is exposed when using TLS

Default: 443

_________
#### CERT_FILE

This value controls what X.509 certificate is offered to clients connecting on the TLS port. The certificate is expected in PEM format. The value is a path name resolved relative to the working directory of Incus.

Default: cert.pem

_________
#### KEY_FILE

This value controls what X.509 private key is used for decrypting the TLS traffic. The key is expected in PEM format.

Default: key.pem

_________
#### APNS_ENABLED

This value controls whether the server will listen for and send iOS push notifications

**false**
> APSN is disabled, the server will not listen for iOS push notifications

**true**
> APNS is enabled, the server will listen for and send iOS push notifications

Default: false

_________
#### APNS_[BUILD]_CERT
Where [BUILD] is one of: DEVELOPMENT, STORE, ENTERPRISE, or BETA

This value controls what APNS granted cert the server will use when calling the APNS API. Each build environment has its own instance of this configuration variable.

Default: myapnsappcert.pem

_________
#### APNS_[BUILD]_PRIVATE_KEY
Where [BUILD] is one of: DEVELOPMENT, STORE, ENTERPRISE, or BETA

This value controls what APNS granted private key the server will use when calling the APNS API. Each build environment has its own instance of this configuration variable.

Default: myapnsappprivatekey.pem

_________
#### APNS_[BUILD]_URL
Where [BUILD] is one of: DEVELOPMENT, STORE, ENTERPRISE, or BETA

This value controls what APNS url is used when calling the APNS API. Each build environment has its own instance of this configuration variable.

Default: gateway.push.apple.com:2195

APNS_DEVELOPMENT_URL defaults to gateway.sandbox.push.apple.com:2195

_________
#### IOS_PUSH_SOUND

This value controls what sound plays on push notification receive.

Default: bingbong.aiff

_________
#### GCM_ENABLED

This value controls whether the server will listen for and send Android push notifications

**false**
> GCM is disabled, the server will not listen for Android push notifications

**true**
> GCM is enabled, the server will listen for and send Android push notifications

Default: false

_________
#### GCM_API_KEY

This is the GCM granted api key used for calling the GCM API.

Default: foobar

_________
#### ANDROID_ERROR_QUEUE

This value controls where Android push errors are stored for later retrieval.

Default: Incus_Android_Error_Queue
