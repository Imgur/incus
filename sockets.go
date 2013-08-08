package main

import (
    "log"
    "net/http"
    "strings"
    "errors"

    "code.google.com/p/go.net/websocket"
)

const listenAddr = "localhost:4000"

func initSocketListener() {
    http.HandleFunc("/", rootHandler)
    http.Handle("/socket", websocket.Handler(Connect))
    
    err := http.ListenAndServe(listenAddr, nil)
    if err != nil {
        log.Fatal(err)
    }
}

type Socket struct {
    ws   *websocket.Conn
    UID  string
    buff chan *Message
    done chan bool
}

type Message struct {
    Name string
    Body map[string]interface{}
    Time int64
}

func (this Message) Handle() {
    log.Printf("Handling message fo type %s\n", this.Name)
    
    if this.Name == "MessageUser" {
        UID, ok := this.Body["UID"].(string)
        if !ok {
            return
        }
        
        rec, err := Store.Client(UID)
        if err != nil {
            return 
        }
        
        rec.buff <- &this
    }
}

func (this Socket) Close() error {
    Store.Remove(this.UID)
    this.done <- true
    
    return nil
}

func Connect(ws *websocket.Conn) {
    sock := Socket{ws, "", make(chan *Message, 1000), make(chan bool)}
    
    log.Printf("Connected via %s\n", ws.RemoteAddr());
    if err := Authenticate(&sock); err != nil {
        log.Printf("Error: %s\n", err.Error())
        return
    }

    go listenForMessages(&sock)
    go listenForWrites(&sock)
    
    <-sock.done
    sock.Close()
}

func Authenticate(sock *Socket) error {
    var message Message
    err := websocket.JSON.Receive(sock.ws, &message)

    log.Println(message.Name)
    if err != nil {
        return err
    }
    
    if strings.ToLower(message.Name) != "handshake" {
        return errors.New("Error: Handshake Expected.\n")
    }
    
    UID, ok := message.Body["UID"].(string)
    if !ok {
        return errors.New("Error on Handshake: Bad Input.\n")
    }
    
    log.Printf("saving UID as %s", UID)
    sock.UID = UID
    Store.Save(UID, sock)
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
                var message Message
                err := websocket.JSON.Receive(sock.ws, &message)
                log.Println("Waiting...\n")
                if err != nil {
                    log.Printf("Error: %s\n", err.Error())
                    
                    sock.Close()
                    return 
                }
                log.Println(message)
                
                message.Handle()
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