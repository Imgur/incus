function Incus(url, UID) {
    this.MAXRETRIES   = 6;
    
    this.retries      = 0;
    this.url          = url;
    this.UID          = UID;
    this.onMessageCbs = {};
    this.connected    = false;
    
    this.connect();
}

Incus.prototype.connect = function() {
    this.socket = new WebSocket(this.url);
    
    var self = this;
    this.socket.onopen    = function() { self.authenticate() };
    this.socket.onmessage = function(e) { self.onMessage(e) };
    this.socket.onclose   = function() { self.onClose() };
}

Incus.prototype.newCommand = function(command, message) {
    message['time'] = Math.round(new Date().getTime() / 1000);
    var obj = {
        "command": command,
        "message": message,
    };
    
    return JSON.stringify(obj);
}

Incus.prototype.authenticate = function() {
    this.retries = 0;
    var message = this.newCommand({'command': "authenticate", 'user': this.UID}, {});
    
    this.socket.send(message);
    console.log("Authenticated");
    
    this.connected = true;
    if("connect" in this.onMessageCbs) {
        this.onMessageCbs["connect"].call(null)
    }
}

Incus.prototype.on = function(name, func) {
    if (name == 'connect' && this.connected) {
        func();
    }
    
    this.onMessageCbs[name] = func;
}

Incus.prototype.onMessage = function(e) {
    var msg = JSON.parse(e.data);

    if ("event" in msg && msg.event in this.onMessageCbs) {
        if(typeof this.onMessageCbs[msg.event] == "function") {
            this.onMessageCbs[msg.event].call(null, msg.data);
        }
    }
}

Incus.prototype.onClose = function() {
    if (this.retries > this.MAXRETRIES) {
        return;
    }
    
    this.retries++;
    this.connected = false;
    
    var self = this;
    window.setTimeout(function() {
        console.log("Connection closed, retrying");
        
        self.connect();
    }, 1000);
}

Incus.prototype.MessageUser = function(event, UID, data) {
    var command = {"command": "message", "user": UID};
    var message = {"event": event, "data": data};
    
    var msg = this.newCommand(command, message);
    return this.socket.send(msg);
}

Incus.prototype.MessagePage = function(event, page, data) {
    var command = {"command": "message", "page": page};
    var message = {"event": event, "data": data};
    
    var msg = this.newCommand(command, message);
    return this.socket.send(msg);
}

Incus.prototype.MessageAll = function(event, data) {
    var command = {"command": "message"};
    var message = {"event": event, "data": data};
    
    var msg = this.newCommand(command, message);
    return this.socket.send(msg);
}

Incus.prototype.setPage = function(page) {
    var command = {'command': 'setpage', 'page': page};
    
    var msg = this.newCommand(command, {});
    return this.socket.send(msg);
}