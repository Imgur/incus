package main

import (
    "net/http"
    "log"
    "time"
    "fmt"
    "os"
    "os/signal"
    "syscall"
    _ "net/http/pprof"
)

var DEBUG bool
var CLIENT_BROAD bool
var store *Storage

func main() {
    store = nil
    signals := make(chan os.Signal, 1)
    
    defer func() {
        if err := recover(); err != nil {
            log.Printf("FATAL: %s", err)
            shutdown()
        }
    }()

    conf  := initConfig()
    initLogger(conf)
    signal.Notify(signals, os.Interrupt, syscall.SIGTERM)
    InstallSignalHandlers(signals)
    
    store = initStore(&conf)

    go func() {
        for {
            log.Println(store.memory.clientCount)
            time.Sleep(20 * time.Second)
        }
    }()
    
    CLIENT_BROAD = conf.GetBool("client_broadcasts")
    server := createServer(&conf, store)
    
    go server.initAppListener()
    go server.initSocketListener()
    go server.initPingListener()
    
    listenAddr := fmt.Sprintf(":%s", conf.Get("listening_port"))
    err := http.ListenAndServe(listenAddr, nil)
    if err != nil {
        log.Fatal(err)
    }
}

func InstallSignalHandlers(signals chan os.Signal) {
    go func() {
        sig := <-signals
        log.Printf("%v caught, incus is going down...", sig)
        shutdown()
    }()
}

func initLogger(conf Configuration) {
    fi, err := os.OpenFile("/var/log/incus.log", os.O_RDWR|os.O_APPEND, 0660);
    if err != nil {
        log.Fatalf("Error: %v", err.Error());
    }
    
    log.SetOutput(fi)
    
    DEBUG = false
    if conf.Get("log_level") == "debug" {
        DEBUG = true
    }
}

func shutdown() {
    if store != nil {
        log.Println("clearing redis memory...")
    }
    
    log.Println("Terminated")
    os.Exit(0)
}