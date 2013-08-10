package main

import (
    "log"
    "encoding/json"
    "errors"
    "time"
)

type Message struct {
    Event string
    Body map[string]interface{}
    Time int64
}

func (this *Message) FromSocket(sock *Socket) {
    log.Printf("Handling message fo type %s\n", this.Event)
    
    switch this.Event {
    case "MessageUser":
        msg, err := this.formatBody()
        if err != nil {
            return
        }
        
        UID, ok := this.Body["UID"].(string)
        if !ok {
            return
        }
        
        rec, err := sock.Server.Store.Client(UID)
        if err != nil {
            return
        }
        
        rec.buff <- msg
    
    case "MessageAll":
        msg_str, _ := json.Marshal(this)
        
        sock.Server.Store.redis.Publish("Message", string(msg_str))
    }
}

func (this *Message) FromRedis(server *Server) {
    log.Printf("Handling message fo type %s\n", this.Event)
    
    switch this.Event {
    
    case "MessageUser":
        msg, err := this.formatBody()
        if err != nil {
            return
        }
        
        UID, ok := this.Body["UID"].(string)
        if !ok {
            return
        }
        
        rec, err := server.Store.Client(UID)
        if err != nil {
            return
        }
        
        rec.buff <- msg
        return
    
    case "MessageAll":
        msg, err := this.formatBody()
        if err != nil {
            return
        }
    
        clients := server.Store.Clients()
        
        for _, sock := range clients {
            sock.buff <- msg
        }
        
        return
    }
}

func (this *Message) formatBody() (*Message, error) {    
    event, e_ok := this.Body["Event"].(string)
    body,  b_ok := this.Body["Message"].(map[string]interface{})
    
    if !b_ok || ! e_ok {
        return nil, errors.New("Could not format message body")
    }
    
    msg := &Message{event, body, time.Now().UTC().Unix()};
    
    return msg, nil
}
