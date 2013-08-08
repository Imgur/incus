package main

import (
    "errors"
 //   "log"
    
   // "github.com/garyburd/redigo/redis"
)

type MemoryStore struct {
    clients     map[string] *Socket
    clientCount int
}

const StorageType = "redis"
var Store = RedisStore{0}

func initStore() {
    if StorageType == "redis" {

        
        Store = RedisStore{0}
    }
}



func (m *MemoryStore) Save(UID string, s *Socket) (bool, error) {
    _, exists := m.clients[UID]
    
    m.clients[UID] = s;
    
    if(!exists) {  // if same UID connects again, don't up the clientCount
        m.clientCount++
    }
    
    return true, nil
}

func (m *MemoryStore) Remove(UID string) (bool, error) {
    _, exists := m.clients[UID]
    
    delete(m.clients, UID)
    
    if(exists) { // only subtract if the client was in the store in the first place.
        m.clientCount--
    }
    
    return true, nil
}

func (m *MemoryStore) Client(UID string) (*Socket, error) {
    var client, exists = m.clients[UID]
    
    if(!exists) {
        return client, errors.New("ClientID doesn't exist")
    }
    return client, nil
}

func (m *MemoryStore) Clients() (map[string] *Socket, error) {
    return m.clients, nil
}

func (m *MemoryStore) Count() (int, error) {
    return m.clientCount, nil
}
