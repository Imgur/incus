package main

import (
    "log"
    "strings"
    "errors"

    "code.google.com/p/go.net/websocket"
)

const listenAddr = "localhost:4000"

type Socket struct {
    ws     *websocket.Conn
    UID    string
    buff   chan *Message
    done   chan bool
    Server *Server
}

func (this Socket) Close() error {
    this.Server.Store.Remove(this.UID)
    this.done <- true
    
    return nil
}

func Authenticate(sock *Socket) error {
    var message Message
    err := websocket.JSON.Receive(sock.ws, &message)

    log.Println(message.Event)
    if err != nil {
        return err
    }
    
    if strings.ToLower(message.Event) != "Authenticate" {
        return errors.New("Error: Authenticate Expected.\n")
    }
    
    UID, ok := message.Body["UID"].(string)
    if !ok {
        return errors.New("Error on Authenticate: Bad Input.\n")
    }
    
    log.Printf("saving UID as %s", UID)
    sock.UID = UID
    sock.Server.Store.Save(UID, sock)
    
    log.Printf("saving UID as %s", sock.UID)
    
    return nil
}

func listenForMessages(sock *Socket) {
    
    for {
        
        select {
            case <- sock.done:
                sock.Close()
                return
            
            default:
                var message *Message
                err := websocket.JSON.Receive(sock.ws, message)
                log.Println("Waiting...\n")
                if err != nil {
                    log.Printf("Error: %s\n", err.Error())
                    
                    sock.Close()
                    return 
                }
                log.Println(message)
                
                go message.FromSocket(sock)
        }
        
    }
}

func listenForWrites(sock *Socket) {
    for {
        select {
            case message := <-sock.buff:
                log.Println("Send:", message)
                if err := websocket.JSON.Send(sock.ws, message); err != nil {
                    sock.Close()
                }
                
            case <-sock.done:
                sock.Close()
                return
        }
    }
}