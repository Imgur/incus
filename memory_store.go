package main

import "errors"

type Page struct {
    clients     map[string] *Socket
    clientCount int64
}

type MemoryStore struct {
    clients     map[string] *Socket
    clientCount int64
    pages       map[string] *Page
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

func (this *MemoryStore) SetPage(UID string, page string) error {
    sock, err := this.Client(UID)
    if err != nil {
        return err
    }
    
    p, exists := this.pages[page]
    if !exists {
        pageMap := make(map[string]*Socket)
        pageMap[UID] = sock
        
        this.pages[page] = &Page{
            pageMap,
            1,
        }
        
        return nil
    }
    
    _, exists = p.clients[UID]
    p.clients[UID] = sock
    if !exists {
        p.clientCount++
    }
    
    return nil
}

func (this *MemoryStore) UnsetPage(UID string, page string) error {
    p, exists := this.pages[page]
    if !exists {
        return nil
    }
    
    _, exists = p.clients[UID]
    delete(p.clients, UID)
    if !exists {
        p.clientCount--
    }
    
    if p.clientCount == 0 {
        delete(this.pages, page)
    }
    
    return nil
}

func (this *MemoryStore) getPage(page string) *Page {
    var p, exists = this.pages[page]
    if !exists {
        return nil
    }
    
    return p
}

func (this *Page) Clients() (map[string] *Socket) {
    return this.clients
}

