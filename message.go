package main

import (
    "log"
)

type Message struct {
    Name string
    Body map[string]interface{}
    Time int64
}

func (this Message) Handle(sock *Socket) {
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
        
        rec.buff <- &this
    }
}