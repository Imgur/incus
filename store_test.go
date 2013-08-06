package main

import "testing"

func TestSave(t *testing.T) {
    Store.Save("TEST", &socket{nil, make(chan bool)})
    
    var _, exists = Store.clients["TEST"]
    if(!exists) {
        t.Errorf("Save function failed, Client not found")
    }
    
    if(Store.clientCount != 1) {
        t.Errorf("Save function failed, clientCount = %v, want %v", Store.clientCount, 1)
    }
    
    Store.Save("TEST1", &socket{nil, make(chan bool)})
    if(Store.clientCount != 2) {
        t.Errorf("Save function failed, clientCount = %v, want %v", Store.clientCount, 2)
    }
    
    Store.Save("TEST1", &socket{nil, make(chan bool)})
    if(Store.clientCount != 2) {
        t.Errorf("Save function failed, clientCount = %v, want %v", Store.clientCount, 2)
    }
    
}
