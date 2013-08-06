package main

import "errors"

type MemoryStore struct {
    clients     map[string] *socket
    clientCount int
}

var Store = MemoryStore{make(map[string]*socket), 0}

func (m *MemoryStore) Save(UID string, s *socket) (bool, error) {
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

func (m *MemoryStore) Client(UID string) (*socket, error) {
    var client, exists = m.clients[UID]
    
    if(!exists) {
        return client, errors.New("ClientID doesn't exist")
    }
    return client, nil
}

func (m *MemoryStore) Clients() (map[string] *socket, error) {
    return m.clients, nil
}

func (m *MemoryStore) Count() (int, error) {
    return m.clientCount, nil
}
