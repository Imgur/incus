package main

import "errors"

type MemoryStore struct {
    clients     map[string] *socket
    clientCount int
}

var Store = MemoryStore{make(map[string]*socket), 0}

func (m *MemoryStore) Save(UID string, s *socket) (bool, error) {
    m.clients[UID] = s;
    m.clientCount++
    
    return true, nil
}

func (m *MemoryStore) Remove(UID string) (bool, error) {
    delete(m.clients, UID)
    m.clientCount--
    
    return true, nil
}

func (m *MemoryStore) GetClient(UID string) (*socket, error) {
    var client, exists = m.clients[UID]
    
    if(!exists) {
        return client, errors.New("ClientID doesn't exist")
    }
    return client, nil
}

func (m *MemoryStore) GetCount(UID string) (int, error) {
    return m.clientCount, nil
}