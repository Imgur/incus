package main

import (
    "log"
    "strings"
    "errors"
    "fmt"

    "code.google.com/p/go.net/websocket"
)

var socketIds chan string
type Socket struct {
    SID    string  // socket ID, randomly generated
    UID    string  // User ID, passed in via client
    Page   string  // Current page, if set.
    
    ws        *websocket.Conn
    Server    *Server
    
    buff      chan *Message
    done      chan bool
    closed    bool
}

func init() {
    socketIds = make(chan string)
    
    go func() {
        var i = 1
        for {
            i++
            socketIds <- fmt.Sprintf("%v", i)
        }
    }()
}

func newSocket(ws *websocket.Conn, server *Server, UID string) *Socket {
    return &Socket{<-socketIds, UID, "", ws, server, make(chan *Message, 1000), make(chan bool), false}
}

func (this *Socket) Close() error {
    if !this.closed {
        this.closed = true
        
        if this.Page != "" {
            this.Server.Store.UnsetPage(this)
            this.Page = ""
        }
        
        this.Server.Store.Remove(this)
        close(this.done)
    }
    
    return nil
}

func (this *Socket) Authenticate() error {
    var message CommandMsg
    err := websocket.JSON.Receive(this.ws, &message)

    if DEBUG { log.Println(message.Command) }
    if err != nil {
        return err
    }
    
    command := message.Command["command"]
    if strings.ToLower(command) != "authenticate" {
        return errors.New("Error: Authenticate Expected.\n")
    }
    
    UID, ok := message.Command["user"]
    if !ok {
        return errors.New("Error on Authenticate: Bad Input.\n")
    }
    
    if DEBUG { log.Printf("saving UID as %s", UID) }
    
    this.UID = UID
    this.Server.Store.Save(this)
        
    return nil
}

func (this *Socket) listenForMessages() {
    for {
        
        select {
            case <- this.done:
                return
            
            default:
                var command CommandMsg
                err := websocket.JSON.Receive(this.ws, &command)
                if err != nil {
                    if DEBUG { log.Printf("Error: %s\n", err.Error()) }
                    
                    go this.Close()
                    return 
                }
                
                if DEBUG { log.Println(command) }
                go command.FromSocket(this)
        }
    }
}

func (this *Socket) listenForWrites() {
    for {
        select {            
            case message := <-this.buff:
                if DEBUG { log.Println("Sending:", message) }
                if err := websocket.JSON.Send(this.ws, message); err != nil {
                    if DEBUG { log.Printf("Error: %s\n", err.Error()) }
                    go this.Close()
                    return
                }
                
            case <-this.done:
                return
        }
    }
}
