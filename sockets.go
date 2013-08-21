package main

import (
    "log"
    "strings"
    "errors"
    "sync"
    "fmt"

    "code.google.com/p/go.net/websocket"
)

var socketIds chan string
type Socket struct {
    SID    string  // socket ID, randomly generated
    UID    string  // User ID, passed in via client
    Page   string  // Current page, if set.
    
    ws     *websocket.Conn
    buff   chan *Message
    done   chan bool
    Server *Server
}

func init() {
    socketIds = make(chan string)
    
    go func() {
        var i = 0
        for {
            i++
            socketIds <- fmt.Sprintf("%v", i)
        }
    }()
}

func newSocket(ws *websocket.Conn, server *Server, UID string) *Socket {
    return &Socket{<-socketIds, UID, "", ws, make(chan *Message, 1000), make(chan bool), server}
}

func (this *Socket) Close() error {
    if DEBUG { log.Printf("CLOSING SOCK %s", this.Page) }
    if this.Page != "" {
        this.Server.Store.UnsetPage(this)
        this.Page = ""
    }
    
    this.Server.Store.Remove(this)
    this.done <- true
    
    return nil
}

func (this *Socket) Authenticate() error {
    var message Message
    err := websocket.JSON.Receive(this.ws, &message)

    if DEBUG { log.Println(message.Event) }
    if err != nil {
        return err
    }
    
    if strings.ToLower(message.Event) != "authenticate" {
        return errors.New("Error: Authenticate Expected.\n")
    }
    
    UID, ok := message.Body["UID"].(string)
    if !ok {
        return errors.New("Error on Authenticate: Bad Input.\n")
    }
    
    if DEBUG { log.Printf("saving UID as %s", UID) }
    
    this.UID = UID
    this.Server.Store.Save(this)
        
    return nil
}

func (this *Socket) listenForMessages(wg *sync.WaitGroup) {
    defer wg.Done()
    for {
        
        select {
            case <- this.done:
                return
            
            default:
                var message Message
                err := websocket.JSON.Receive(this.ws, &message)
                
                if err != nil {
                    if DEBUG { log.Printf("Error: %s\n", err.Error()) }
                    
                    this.Close()
                    return 
                }
                if DEBUG { log.Println(message) }
                
                go message.FromSocket(this)
        }
        
    }
}

func (this *Socket) listenForWrites(wg *sync.WaitGroup) {
    defer wg.Done()
    for {
        select {
            case message := <-this.buff:
                if DEBUG { log.Println("Sending:", message) }
                if err := websocket.JSON.Send(this.ws, message); err != nil {
                    if DEBUG { log.Printf("Error: %s\n", err.Error()) }
                    this.Close()
                    return
                }
                
            case <-this.done:
                return
        }
    }
}
