package main

import (
    //"fmt"
    
)

type logger struct {
    log   chan string
    error chan string
    print chan string
}

//var Logger     = logger{ make(chan string), make(chan string), make(chan error) }
var errorLevel = "debug"

func initLogger() error {
    //if(errorLevel == "debug") {
    //    go LoggerListen()
    //}
    return nil
}

//func LoggerListen() {
//    //for {
//        var message <- Logger.log
//        
//        fmt.Println(message)
//    //}
//}