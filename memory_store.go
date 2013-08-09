package main

import "errors"

type MemoryStore struct {
    clients     map[string] *Socket
    clientCount int64
}

func (this *MemoryStore) Save(UID string, s *Socket) (error) {
    _, exists := this.clients[UID]
    
    this.clients[UID] = s;
    
    if(!exists) {  // if same UID connects again, don't up the clientCount
        this.clientCount++
    }
    
    return nil
}

func (this *MemoryStore) Remove(UID string) (bool, error) {
    _, exists := this.clients[UID]
    
    delete(this.clients, UID)
    
    if(exists) { // only subtract if the client was in the store in the first place.
        this.clientCount--
    }
    
    return true, nil
}

func (this *MemoryStore) Client(UID string) (*Socket, error) {
    var client, exists = this.clients[UID]
    
    if(!exists) {
        return nil, errors.New("ClientID doesn't exist")
    }
    return client, nil
}

func (this *MemoryStore) Clients() (map[string] *Socket) {
    return this.clients
}

func (this *MemoryStore) Count() (int64, error) {
    return this.clientCount, nil
}
