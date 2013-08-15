package main

import (
    "testing"
    "time"
    "net/http"
    
    "code.google.com/p/go.net/websocket"
)

var server = startTestServer()

func startTestServer() *Server {
   var Store =  initStore(nil)
   var server = createServer(nil, &Store)

   go server.initSocketListener()
 
   go http.ListenAndServe(listenAddr, nil)

   return server
}


func TestAuthenticate(t *testing.T) {
    ws, err := websocket.Dial("ws://localhost:4000/socket", "", "http://localhost/")
    if err != nil {
        t.Fatalf("Authenticate Test failed, Could not connect with error %s", err.Error())
    }
    
    //sock       := Socket{ws, "", make(chan *Message), make(chan bool)}
    body       := make(map[string]interface{})
    body["UID"] = "TEST_UID"
    message    := Message{"Authenticate", body, 1234}
    
    if err := websocket.JSON.Send(ws, message); err != nil {
        t.Fatal("Authenticate Test failed, Could not send Auth message")
    }

    time.Sleep(10 * time.Millisecond)
    
    _, err1 := server.Store.Client("TEST_UID")
    
    if err1 != nil {
        t.Errorf("Authenticate Test failed, Failed to Authenticate expected TEST_UID, got %s", err1.Error())
    }    
}

func TestListenForMessages(t *testing.T) {
    sock, err := server.Store.Client("TEST_UID")
    if err != nil {
        t.Fatalf("ListenForMessages Test failed, Failed to get Socket TEST_UID, with error: %s", err.Error())
    }
    
    body       := make(map[string]interface{})
    body["UID"] = "TEST_UID"
    message    := Message{"MessageUser", body, 1234}
    
    if err = websocket.JSON.Send(sock.ws, message); err != nil {
        t.Errorf("ListenForMessages Test failed, Failed to get Socket TEST_UID, with error: %s", err.Error())
    }
}

func TestListenForWriter(t *testing.T) {
    t.Fatal("somthing is wrong with this stupes test") //CLIENT WS VS. SERVER WS!?
    sock, err := server.Store.Client("TEST_UID")
    if err != nil {
        t.Fatalf("ListenForWriter Test failed, Failed to get Socket TEST_UID, with error: %s", err.Error())
    }
    
    go func() {
        time.Sleep(10 * time.Millisecond)
        body       := make(map[string]interface{})
        body["UID"]     = "TEST_UID"
        msg := make(map[string]int)
        msg["t"] = 1
        body["Message"] = msg
        message    := Message{"MessageUser", body, 1234}

        sock.buff <- &message
    }()
    
    var message Message
    if err = websocket.JSON.Receive(sock.ws, &message); err != nil {
        t.Errorf("ListenForMessages Test failed, Failed to get Socket TEST_UID, with error: %s", err.Error())
    }
    
    if message.Event != "MessageUser" {
        t.Error("ListenForMessages Test failed, Failed to get Socket TEST_UID")
    }
    sock.Close()
}
