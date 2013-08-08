package main

import (
    "net/http"
    "log"

    "code.google.com/p/go.net/websocket"
)

type Server struct {
    Config  *Configuration
    Store   *Storage
}

func main() {
    conf  := initConfig()
    store := initStore(&conf)
    //initLogger()
    server := Server{&conf, &store}
    
    //server.initAppListner()
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
    
    http.Handle("/socket", websocket.Handler(Connect))
}