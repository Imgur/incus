package incus

import "errors"

type MemoryStore struct {
	clients map[string]map[string]*Socket
	pages   map[string]map[string]*Socket
	groups  map[string]map[string]*Socket

	clientCount int64
}

func (this *MemoryStore) Save(sock *Socket) error {
	user, exists := this.clients[sock.UID]

	if !exists {
		this.clientCount++

		userMap := make(map[string]*Socket)
		userMap[sock.SID] = sock
		this.clients[sock.UID] = userMap

		return nil
	}

	_, exists = user[sock.SID]
	user[sock.SID] = sock
	if !exists {
		this.clientCount++
	}

	return nil
}

func (this *MemoryStore) Remove(sock *Socket) error {
	user, exists := this.clients[sock.UID]
	if !exists { // only subtract if the client was in the store in the first place.
		return nil
	}

	_, exists = user[sock.SID]
	delete(user, sock.SID)
	if exists {
		this.clientCount--
	}

	if len(user) == 0 {
		delete(this.clients, sock.UID)
	}

	return nil
}

func (this *MemoryStore) Client(UID string) (map[string]*Socket, error) {
	var client, exists = this.clients[UID]

	if !exists {
		return nil, errors.New("ClientID doesn't exist")
	}
	return client, nil
}

func (this *MemoryStore) Clients() map[string]map[string]*Socket {
	return this.clients
}

func (this *MemoryStore) Count() (int64, error) {
	return this.clientCount, nil
}

func (this *MemoryStore) SetPage(sock *Socket) error {
	page, exists := this.pages[sock.Page]
	if !exists {
		pageMap := make(map[string]*Socket)
		pageMap[sock.SID] = sock
		this.pages[sock.Page] = pageMap

		return nil
	}

	page[sock.SID] = sock

	return nil
}

func (this *MemoryStore) UnsetPage(sock *Socket) error {
	page, exists := this.pages[sock.Page]
	if !exists {
		return nil
	}

	delete(page, sock.SID)

	if len(page) == 0 {
		delete(this.pages, sock.Page)
	}

	return nil
}

func (this *MemoryStore) getPage(page string) map[string]*Socket {
	var p, exists = this.pages[page]

	if !exists {
		return nil
	}

	return p
}

func (this *MemoryStore) SetGroup(sock *Socket) error {
	group, exists := this.groups[sock.Page]
	if !exists {
		groupMap := make(map[string]*Socket)
		groupMap[sock.SID] = sock
		this.groups[sock.Group] = groupMap

		return nil
	}

	group[sock.SID] = sock

	return nil
}

func (this *MemoryStore) UnsetGroup(sock *Socket) error {
	group, exists := this.groups[sock.Group]
	if !exists {
		return nil
	}

	delete(group, sock.SID)

	if len(group) == 0 {
		delete(this.groups, sock.Group)
	}

	return nil
}

func (this *MemoryStore) getGroup(group string) map[string]*Socket {
	var g, exists = this.groups[group]

	if !exists {
		return nil
	}

	return g
}
