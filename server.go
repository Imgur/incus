package main

import (
    "net/http"
    "log"
    "encoding/json"
    "crypto/md5"
    "time"
    "io"
    "fmt"
    "sync"

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

func (this *Server) initSocketListener() {
    Connect := func(ws *websocket.Conn) {
        sock := newSocket(ws, this, "")
        
        if DEBUG { log.Printf("Socket connected via %s\n", ws.RemoteAddr()) }
        if err := sock.Authenticate(); err != nil {
            if DEBUG { log.Printf("Error: %s\n", err.Error()) }
            return
        }
    
        var wg sync.WaitGroup
        wg.Add(2)
        
        go sock.listenForMessages(&wg)
        go sock.listenForWrites(&wg)
        
        wg.Wait()
        if DEBUG { log.Println("Socket Closed") }
    }
    
    http.Handle("/socket", websocket.Handler(Connect))
}

func (this *Server) initAppListener() {
    rec := make(chan []string)
    
    consumer, err := this.Store.redis.Subscribe(rec, "Message")
    if err != nil {
        log.Fatal("Couldn't subscribe to redis channel")
    }
    defer consumer.Quit()
    
    if DEBUG { log.Println("LISENING FOR REDIS MESSAGE") }
    var ms []string
    for {
        var msg Message
        ms = <- rec
        json.Unmarshal([]byte(ms[2]), &msg)
        go msg.FromRedis(this)
        
        if DEBUG { log.Printf("Received %v\n", msg.Event) }
    }  
}

func (this *Server) initPingListener() {
    pingHandler := func(w http.ResponseWriter, r *http.Request) {
        fmt.Fprint(w, "OK")
    }   
    
    http.HandleFunc("/ping", pingHandler)
}