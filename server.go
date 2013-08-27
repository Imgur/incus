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
    _ "net/http/pprof"

    "code.google.com/p/go.net/websocket"
)

var DEBUG bool
var CLIENT_BROAD bool

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
    defer func() {
        if err := recover(); err != nil {
            log.Println("FATAL: ", err)

            log.Println("clearing redis memory")
            log.Println("Shutting down")
        }
    }()

    conf  := initConfig()
    store := initStore(&conf)
    initLogger(conf)
    
    CLIENT_BROAD = conf.GetBool("client_broadcasts")
    server := createServer(&conf, &store)
    
    go server.initAppListner()
    go server.initSocketListener()
    
    listenAddr := fmt.Sprintf(":%s", conf.Get("listening_port"))
    http.Handle("/js/", http.StripPrefix("/js/", http.FileServer(http.Dir("./js/"))))
    err := http.ListenAndServe(listenAddr, nil)
    if err != nil {
        log.Fatal(err)
    }
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

func (this *Server) initAppListner() {
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
    
