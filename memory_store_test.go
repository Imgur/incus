package main

import "testing"

var MemStore = MemoryStore{make(map[string]*Socket), 0}

func makeSock(UID string) *Socket {
    return &Socket{nil, UID, make(chan *Message), make(chan bool), nil}
}

func TestSave(t *testing.T) {
    MemStore.Save("TEST", makeSock("TEST"))
    
    _, exists := MemStore.clients["TEST"]
    if(!exists) {
        t.Errorf("Save Test failed, Client not found")
    }
    
    if(MemStore.clientCount != 1) {
        t.Errorf("Save Test failed, clientCount = %v, want %v", MemStore.clientCount, 1)
    }
    
    MemStore.Save("TEST1", makeSock("TEST1"))
    if(MemStore.clientCount != 2) {
        t.Errorf("Save Test failed, clientCount = %v, want %v", MemStore.clientCount, 2)
    }
    
    MemStore.Save("TEST1", makeSock("TEST1"))
    if(MemStore.clientCount != 2) {
        t.Errorf("Save Test failed, clientCount = %v, want %v", MemStore.clientCount, 2)
    }
}

func TestRemove(t *testing.T) {
    if(MemStore.clientCount != 2) {
        t.Errorf("Remove Test is invalid, clientCount = %v, want %v", MemStore.clientCount, 2)
    }
    
    MemStore.Remove("TEST")
    _, exists := MemStore.clients["TEST"];
    
    if exists {
        t.Errorf("Remove Test failed, Client was not removed")
    }
    
    if(MemStore.clientCount != 1) {
        t.Errorf("Remove Test failed, clientCount = %v, want %v", MemStore.clientCount, 1)
    }
    
    MemStore.Remove("TEST")
    if(MemStore.clientCount != 1) {
        t.Errorf("Remove Test failed, clientCount = %v, want %v", MemStore.clientCount, 1)
    }
    
    MemStore.Remove("TEST1")
    if(MemStore.clientCount != 0) {
        t.Errorf("Remove Test failed, clientCount = %v, want %v", MemStore.clientCount, 0)
    }
    
    if len(MemStore.clients) != 0 {
        t.Errorf("Remove Test failed, clients map expected to be empty")
    }
}

func TestClient(t *testing.T) {
    MemStore.Save("TEST1", makeSock("TEST1"))
    
    client, err := MemStore.Client("TEST1");
    if err != nil {
        t.Errorf("Client Test failed, client TEST1 should exist")
    }
    go func() { client.done <- true }()
    val := <- client.done
    
    if(val != true) {
        t.Errorf("Client Test failed, could not access client TEST1's data")
    }

    _, err1 := MemStore.Client("N/A");
    if err1 == nil {
        t.Errorf("GetClient Test failed, non-existant client failed to throw error")
    }
    MemStore.Remove("TEST1")
}

func TestGetCount(t *testing.T) {
    MemStore.Save("TEST3", makeSock("TEST3"))

    count, _ := MemStore.Count()
    if count != 1 {
        t.Errorf("GetCount Test failed. ClientCount = %v, expected %v", count, 1)
    }

    MemStore.Save("TEST4", makeSock("TEST4"))
    count, _ = MemStore.Count()
    if count != 2 {
        t.Errorf("GetCount Test failed. ClientCount = %v, expected %v", count, 2 )
    }

}
