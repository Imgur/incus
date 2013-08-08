package main

import "testing"

func TestSave(t *testing.T) {
    Store.Save("TEST", &Socket{nil, "TEST", make(chan *Message), make(chan bool)})
    
    _, exists := Store.clients["TEST"]
    if(!exists) {
        t.Errorf("Save Test failed, Client not found")
    }
    
    if(Store.clientCount != 1) {
        t.Errorf("Save Test failed, clientCount = %v, want %v", Store.clientCount, 1)
    }
    
    Store.Save("TEST1", &Socket{nil, "TEST1", make(chan *Message), make(chan bool)})
    if(Store.clientCount != 2) {
        t.Errorf("Save Test failed, clientCount = %v, want %v", Store.clientCount, 2)
    }
    
    Store.Save("TEST1", &Socket{nil, "TEST1", make(chan *Message), make(chan bool)})
    if(Store.clientCount != 2) {
        t.Errorf("Save Test failed, clientCount = %v, want %v", Store.clientCount, 2)
    }
}

func TestRemove(t *testing.T) {
    if(Store.clientCount != 2) {
        t.Errorf("Remove Test is invalid, clientCount = %v, want %v", Store.clientCount, 2)
    }
    
    Store.Remove("TEST")
    _, exists := Store.clients["TEST"];
    
    if exists {
        t.Errorf("Remove Test failed, Client was not removed")
    }
    
    if(Store.clientCount != 1) {
        t.Errorf("Remove Test failed, clientCount = %v, want %v", Store.clientCount, 1)
    }
    
    Store.Remove("TEST")
    if(Store.clientCount != 1) {
        t.Errorf("Remove Test failed, clientCount = %v, want %v", Store.clientCount, 1)
    }
    
    Store.Remove("TEST1")
    if(Store.clientCount != 0) {
        t.Errorf("Remove Test failed, clientCount = %v, want %v", Store.clientCount, 0)
    }
    
    if len(Store.clients) != 0 {
        t.Errorf("Remove Test failed, clients map expected to be empty")
    }
}

func TestClient(t *testing.T) {
    Store.Save("TEST1", &Socket{nil, "TEST1", make(chan *Message), make(chan bool)})
    
    client, err := Store.Client("TEST1");
    if err != nil {
        t.Errorf("Client Test failed, client TEST1 should exist")
    }
    go func() { client.done <- true }()
    val := <- client.done
    
    if(val != true) {
        t.Errorf("Client Test failed, could not access client TEST1's data")
    }

    _, err1 := Store.Client("N/A");
    if err1 == nil {
        t.Errorf("GetClient Test failed, non-existant client failed to throw error")
    }
    Store.Remove("TEST1")
}

func TestGetCount(t *testing.T) {
    Store.Save("TEST3", &Socket{nil, "TEST1", make(chan *Message), make(chan bool)})

    count, _ := Store.Count()
    if count != 1 {
        t.Errorf("GetCount Test failed. ClientCount = %v, expected %v", count, 1)
    }

    Store.Save("TEST4", &Socket{nil, "TEST1", make(chan *Message), make(chan bool)})
    count, _ = Store.Count()
    if count != 2 {
        t.Errorf("GetCount Test failed. ClientCount = %v, expected %v", count, 2 )
    }

}
