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
    log.Printf("Handling message of type %s\n", this.Event)
    
    switch this.Event {
    case "MessageUser":
        if(sock.Server.Store.StorageType == "redis") {
            this.forwardToRedis(sock.Server)
            return 
        }
        
        this.messageUser(sock.Server)
        
    case "MessageAll":
        if(sock.Server.Store.StorageType == "redis") {
            this.forwardToRedis(sock.Server)
            return
        }
        
        this.messageAll(sock.Server)
    case "SetPage":
        page, ok := this.Body["Page"].(string)
        if !ok {
            return
        }
       log.Println(page) 
        if sock.Page != "" {
            sock.Server.Store.UnsetPage(sock.UID, sock.Page)  //remove old page if it exists
        }
        
        sock.Page = page
        sock.Server.Store.SetPage(sock.UID, page) // set new page
    }
}

func (this *Message) FromRedis(server *Server) {
    log.Printf("Handling message of type %s\n", this.Event)
    
    switch this.Event {
    
    case "MessageUser":
        this.messageUser(server)
    
    case "MessageAll":
        this.messageAll(server)

    case "MessagePage": 
        msg, err := this.formatBody()
        if err != nil {
            return
        }

        page, ok := this.Body["Page"].(string)
        if !ok {
            return
        }

        pageStruct := server.Store.getPage(page)
        if pageStruct == nil {
            return
        }
        
        clients := pageStruct.Clients() 
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

func (this *Message) messageUser(server *Server) {
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
}

func (this *Message) messageAll(server *Server) {
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

func (this *Message) forwardToRedis(server *Server) {
    msg_str, _ := json.Marshal(this)
    server.Store.redis.Publish("Message", string(msg_str)) //pass the message into redis to send message across cluster    
}

