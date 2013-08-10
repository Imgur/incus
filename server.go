package main

import (
    "net/http"
    "log"
    "encoding/json"
    "crypto/md5"
    "time"
    "io"

    "code.google.com/p/go.net/websocket"
)

type Server struct {
    ID      string
    Config  *Configuration
    Store   *Storage
}

func createServer(conf *Configuration, store *Storage) *Server{
    hash := md5.New()
    io.WriteString(hash, time.Now().String())
    id := string(hash.Sum(nil))
    
    return &Server{id, conf, store}
}

func main() {
    conf  := initConfig()
    store := initStore(&conf)
    //initLogger()
    server := createServer(&conf, &store)
    
    go server.initAppListner()
    go server.initSocketListener()
    
    http.HandleFunc("/", rootHandler)
    err := http.ListenAndServe(listenAddr, nil)
    if err != nil {
        log.Fatal(err)
    }
}

func (this *Server) initSocketListener() {
    Connect := func(ws *websocket.Conn) {
        sock := Socket{ws, "", make(chan *Message, 1000), make(chan bool), this}
        defer sock.Close()
        
        log.Printf("Connected via %s\n", ws.RemoteAddr());
        if err := Authenticate(&sock); err != nil {
            log.Printf("Error: %s\n", err.Error())
            return
        }
    
        go listenForMessages(&sock)
        go listenForWrites(&sock)
        
        <-sock.done
    }
    
    http.Handle("/socket", websocket.Handler(Connect))
}

func (this *Server) initAppListner() {
    rec := make(chan []string)
    
    consumer, err := this.Store.redis.Subscribe(rec, "Message")
    if err != nil {
        log.Fatal("Couldn't subscribe to redis channel")
    }
    defer consumer.Quit()
    <- rec // ignore subscribe command
    
    var ms []string
    for {
        var msg Message
        log.Println("LISENING FOR REDIS MESSAGE")
        ms = <- rec
        json.Unmarshal([]byte(ms[2]), &msg)
        log.Printf("Received %v\n", msg.Name)
        
        go msg.FromRedis(this)
    }  
}
    