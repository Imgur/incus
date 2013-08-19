package main

import (
    "os"
    "log"
    
)

type logger struct {
    log   chan string
    error chan string
    print chan string
}

//var Logger     = logger{ make(chan string), make(chan string), make(chan error) }
var errorLevel = "debug"

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

//func LoggerListen() {
//    //for {
//        var message <- Logger.log
//        
//        fmt.Println(message)
//    //}
//}