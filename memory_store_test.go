package incus

import (
	"testing"
)

var Stats = &DiscardStats{}
var ConfigFilePath = "./"
var MemStore = NewStore(Stats).memory
var Socket1 *Socket
var Socket2 *Socket
var Socket3 *Socket

func init() {
	Socket1 = newSocket(nil, nil, nil, "TEST")
	Socket2 = newSocket(nil, nil, nil, "TEST1")
	Socket3 = newSocket(nil, nil, nil, "TEST1")

	NewConfig(ConfigFilePath)
}

func TestSave(t *testing.T) {
	MemStore.Save(Socket1)

	_, exists := MemStore.clients["TEST"]
	if !exists {
		t.Errorf("Save Test failed, Client not found")
	}

	if MemStore.clientCount != 1 {
		t.Errorf("Save Test failed, clientCount = %v, want %v", MemStore.clientCount, 1)
	}

	MemStore.Save(Socket2)
	if MemStore.clientCount != 2 {
		t.Errorf("Save Test failed, clientCount = %v, want %v", MemStore.clientCount, 2)
	}

	MemStore.Save(Socket3)
	if MemStore.clientCount != 3 {
		t.Errorf("Save Test failed, clientCount = %v, want %v", MemStore.clientCount, 3)
	}
}

func TestRemove(t *testing.T) {
	if MemStore.clientCount != 3 {
		t.Errorf("Remove Test is invalid, clientCount = %v, want %v", MemStore.clientCount, 3)
	}

	MemStore.Remove(Socket1)
	_, exists := MemStore.clients["TEST"]

	if exists {
		t.Errorf("Remove Test failed, Client was not removed")
	}

	if MemStore.clientCount != 2 {
		t.Errorf("Remove Test failed, clientCount = %v, want %v", MemStore.clientCount, 2)
	}

	MemStore.Remove(Socket1)
	if MemStore.clientCount != 2 {
		t.Errorf("Remove Test failed, clientCount = %v, want %v", MemStore.clientCount, 2)
	}

	MemStore.Remove(Socket2)
	if MemStore.clientCount != 1 {
		t.Errorf("Remove Test failed, clientCount = %v, want %v", MemStore.clientCount, 1)
	}

	if len(MemStore.clients) != 1 {
		t.Errorf("Remove Test failed, clients map expected to be empty")
	}

	MemStore.Remove(Socket3)
	if MemStore.clientCount != 0 {
		t.Errorf("Remove Test failed, clientCount = %v, want %v", MemStore.clientCount, 0)
	}

	if len(MemStore.clients) != 0 {
		t.Errorf("Remove Test failed, clients map expected to be empty")
	}
}

func TestClient(t *testing.T) {
	MemStore.Save(Socket2)

	client, err := MemStore.Client("TEST1")
	if err != nil {
		t.Errorf("Client Test failed, client TEST1 should exist")
	}

	if client[Socket2.SID].UID != "TEST1" {
		t.Errorf("Client Test failed, could not access client TEST1's data")
	}

	_, err1 := MemStore.Client("N/A")
	if err1 == nil {
		t.Errorf("GetClient Test failed, non-existant client failed to throw error")
	}
	MemStore.Remove(Socket2)
}

func TestGetCount(t *testing.T) {
	MemStore.Save(Socket1)

	count, _ := MemStore.Count()
	if count != 1 {
		t.Errorf("GetCount Test failed. ClientCount = %v, expected %v", count, 1)
	}

	MemStore.Save(Socket2)
	count, _ = MemStore.Count()
	if count != 2 {
		t.Errorf("GetCount Test failed. ClientCount = %v, expected %v", count, 2)
	}

}
