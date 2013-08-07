package main

import (
    "testing"
    "time"
    //"net/http"
    
    "code.google.com/p/go.net/websocket"
)

func TestInitSocketListener(t *testing.T) {
    go initSocketListener()
    time.Sleep(10 * time.Millisecond)
    ws, err := websocket.Dial("ws://localhost:4000/socket", "", "http://localhost/")
    if err != nil {
        t.Errorf("initSocketListener Test failed, Could not connect")
    }
    
    ws.Close()
    
}

func TestAuthenticate(t *testing.T) {

    ws, err := websocket.Dial("ws://localhost:4000/socket", "", "http://localhost/")
    if err != nil {
        t.Fatalf("Authenticate Test failed, Could not connect with error %s", err.Error())
    }
    
    //sock       := Socket{ws, "", make(chan *Message), make(chan bool)}
    body       := make(map[string]interface{})
    body["UID"] = "TEST_UID"
    message    := Message{"Handshake", body, 1234}
    
    if err := websocket.JSON.Send(ws, message); err != nil {
        t.Fatal("Authenticate Test failed, Could not send Auth message")
    }

    time.Sleep(10 * time.Millisecond)
    
    _, err1 := Store.Client("TEST_UID")
    
    if err1 != nil {
        t.Errorf("Authenticate Test failed, Failed to Authenticate expected TEST_UID, got %s", err1.Error())
    }    
}

func TestListenForMessages(t *testing.T) {
    sock, err := Store.Client("TEST_UID")
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
    sock, err := Store.Client("TEST_UID")
    if err != nil {
        t.Fatalf("ListenForWriter Test failed, Failed to get Socket TEST_UID, with error: %s", err.Error())
    }
    
    go func() {
        time.Sleep(100 * time.Millisecond)
        body       := make(map[string]interface{})
        body["UID"] = "TEST_UID"
        message    := Message{"MessageUser", body, 1234}
        
        sock.buff <- &message
    }()
    
    var message Message
    t.Log("WWWWWWWWW")
    if err = websocket.JSON.Receive(sock.ws, &message); err != nil {
        t.Errorf("ListenForMessages Test failed, Failed to get Socket TEST_UID, with error: %s", err.Error())
    }
    
    if message.Name != "MessageUser" {
        t.Error("ListenForMessages Test failed, Failed to get Socket TEST_UID")
    }
    sock.Close()
}
