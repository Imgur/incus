package main

import (
    "log"
    "encoding/json"
)

type Message struct {
    Name string
    Body map[string]interface{}
    Time int64
}

func (this *Message) FromSocket(sock *Socket) {
    log.Printf("Handling message fo type %s\n", this.Name)
    
    if this.Name == "MessageUser" {
        UID, ok := this.Body["UID"].(string)
        if !ok {
            return
        }
        
        rec, err := sock.Server.Store.Client(UID)
        if err != nil {
            return
        }
        
        rec.buff <- this
    }
    
    if this.Name == "MessageAll" {
        msg_str, _ := json.Marshal(this)
        
        sock.Server.Store.redis.Publish("Message", string(msg_str))
    }
}

func (this *Message) FromRedis(server *Server) {
    log.Printf("Handling message fo type %s\n", this.Name)
    
    switch this.Name {
    
    case "MessageUser":
        UID, ok := this.Body["UID"].(string)
        if !ok {
            return
        }
        
        rec, err := server.Store.Client(UID)
        if err != nil {
            return
        }
        
        rec.buff <- this
        return
    
    case "MessageAll":
        clients := server.Store.Clients()
        
        for _, sock := range clients {
            sock.buff <- this
        }
        
        return
    }
}
