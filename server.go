package main

import (
    "net/http"
    "log"
    "encoding/json"
    "crypto/md5"
    "time"
    "io"
    "fmt"

    "code.google.com/p/go.net/websocket"
)

type Server struct {
    ID      string
    Config  *Configuration
    Store   *Storage
    
    timeout time.Duration
}

func createServer(conf *Configuration, store *Storage) *Server{
    hash := md5.New()
    io.WriteString(hash, time.Now().String())
    id := string(hash.Sum(nil))
    
    timeout := time.Duration(conf.GetInt("connection_timeout"))
    return &Server{id, conf, store, timeout}
}

func (this *Server) initSocketListener() {
    Connect := func(ws *websocket.Conn) {
        defer func() { if DEBUG { log.Println("Socket Closed") } }()
        
        sock := newSocket(ws, this, "")
        
        if DEBUG { log.Printf("Socket connected via %s\n", ws.RemoteAddr()) }
        if err := sock.Authenticate(); err != nil {
            if DEBUG { log.Printf("Error: %s\n", err.Error()) }
            return
        }

        go sock.listenForMessages()
        go sock.listenForWrites()
        
        if this.timeout <= 0 { // if timeout is 0 then wait forever and return when socket is done.
            <-sock.done
            return 
        }
        
        select{
            case <- time.After(this.timeout * time.Second):
                sock.Close()
                return
            case <- sock.done:
                return
        }
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
        var cmd CommandMsg
        ms = <- rec
        json.Unmarshal([]byte(ms[2]), &cmd)
        go cmd.FromRedis(this)
    }  
}

func (this *Server) initPingListener() {
    pingHandler := func(w http.ResponseWriter, r *http.Request) {
        fmt.Fprint(w, "OK")
    }   
    
    http.HandleFunc("/ping", pingHandler)
}