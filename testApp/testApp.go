package main

import (
    "log"
    "time"
    "fmt"
    "math/rand"
    
    "menteslibres.net/gosexy/redis"
    "code.google.com/p/go.net/websocket"
)


func main() {
//    pageSample()
//mapsSample()
    pageUpdates()
}

func random(min, max int) int {
    //rand.Seed(time.Now().Nanosecond())
    return rand.Intn(max - min) + min
}

func mapsSample() {
    client := redis.New()
    err := client.Connect("localhost", 6379)

    if err != nil {
        log.Fatalf("Connect failed: %s\n", err.Error())
        return
    }
    
    var message string
    for {
        message =  fmt.Sprintf("{\"Event\":\"MessageAll\",\"Body\":{\"Event\": \"tweet\", \"Message\": {\"coordinates\": [%v, %v]}},\"Time\":12312}", (random(0, 123)*-1), random(0, 64))
        client.Publish("Message", message)
        
        time.Sleep(20 * time.Millisecond)
        
    }
}

var pages = []string{"image", "new", "popular", "realtime", "memes", "gifs", "cats", "dogs", "seals", "headphones"}
func pageSample() {
    shut := make(chan bool)
    store := []*websocket.Conn{}
    
    for i := 0; i < 1020; i++ {
        
            ws, err := websocket.Dial("ws://localhost:4000/socket", "", "http://localhost/")
            if err != nil {
                log.Printf("initSocketListener Test failed, Could not connect")
            }
            
            Auth := fmt.Sprintf("{\"Event\":\"Authenticate\",\"Body\":{\"UID\":\"%v\"},\"Time\":1376138699}", random(0, 100000))
            if err := websocket.Message.Send(ws, Auth); err != nil {
                log.Fatal("Authenticate Test failed, Could not send Auth message")
            }
            
            page := fmt.Sprintf("{\"Event\":\"SetPage\",\"Body\":{\"Page\":\"%s\"},\"Time\":1376138699}", pages[random(0, len(pages))])
            if err := websocket.Message.Send(ws, page); err != nil {
                log.Fatal("Authenticate Test failed, Could not send Auth message")
            }
            
            store = append(store, ws)
        time.Sleep(20 * time.Millisecond)
    }
    
                <-shut
}

func pageUpdates() {
    client := redis.New()
    err := client.Connect("localhost", 6379)
         
    if err != nil {
        log.Fatalf("Connect failed: %s\n", err.Error())
        return
    }   
        
    var message string
    for {
        page := pages[random(0, len(pages))]
        message =  fmt.Sprintf("{\"Event\":\"MessagePage\",\"Body\":{\"Event\": \"update\", \"Page\": \"%s\", \"Message\": {\"page\": \"%s\"}}, \"Time\":12312}",  page, page)
        client.Publish("Message", message)
        time.Sleep(20 * time.Millisecond)
               
    }
}
