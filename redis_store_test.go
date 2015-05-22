package main

import (
	"testing"
	//"time"

	//"github.com/gosexy/redis"
)

//var redisStore = makeTestStore()

// func makeTestStore() RedisStore {
// 	var redisStore = initStore(nil).redis
// 	redisStore.clientsKey = "TestClientKey"
// 	redisStore.pool.maxIdle = 1
// 	return redisStore
// }

func TestPooltestConn(t *testing.T) {
	// conn := redis.New()
	// err := conn.Connect("localhost", 6379)
	// err = redisStore.pool.testConn(conn)
	// if err != nil {
	// 	t.Fatalf("testConn test failed: %s", err.Error())
	// }

	// conn.Quit()
	// err = redisStore.pool.testConn(conn)
	// if err == nil {
	// 	t.Fatalf("testConn test failed: expected error got nil")
	// }
	return
}

func TestPoolClose(t *testing.T) {
	// conn := redis.New()
	// err := conn.Connect("localhost", 6379)
	// if err != nil {
	// 	t.Errorf("pool.Close test failed: could not get connection")
	// }

	// if len(redisStore.pool.connections) != 0 {
	// 	t.Errorf("pool.Close test failed: connection pool not empty")
	// }
	// redisStore.pool.Close(conn)

	// if len(redisStore.pool.connections) != 1 {
	// 	t.Errorf("pool.Close test failed: connection pool length expected to be 1 got %v", len(redisStore.pool.connections))
	// }

	// conn = redis.New()
	// err = conn.Connect("localhost", 6379)
	// if err != nil {
	// 	t.Errorf("pool.Close test failed: could not get connection")
	// }

	// redisStore.pool.Close(conn)
	// if len(redisStore.pool.connections) != 1 { //testing maxIdle; maxIdle = 1
	// 	t.Errorf("pool.Close test failed: connection pool length expected to be 1 got %s", len(redisStore.pool.connections))
	// }
	return
}

func TestPoolGet(t *testing.T) {
	// if len(redisStore.pool.connections) != 1 {
	// 	t.Fatalf("pool.Get test failed: connection pool length expected to be 1 got %s", len(redisStore.pool.connections))
	// }

	// client, ok := redisStore.pool.Get() // retrieve the connection that we created in the previous test
	// if !ok {
	// 	t.Fatalf("pool.Close test failed: could not get connection")
	// }
	// if len(redisStore.pool.connections) != 0 {
	// 	t.Errorf("pool.Get test failed: connection pool length expected to be 0 got %s", len(redisStore.pool.connections))
	// }
	// client.Quit()

	// client, ok = redisStore.pool.Get() // create a new connection using connFn
	// if !ok {
	// 	t.Fatalf("pool.Close test failed: could not get connection")
	// }
	// if _, err := client.Ping(); err != nil {
	// 	t.Errorf("pool.Close test failed: could not ping new connection")
	// }
	// client.Quit()
}

func TestSubscribeAndPublish(t *testing.T) {
	// rec := make(chan []string)
	// conn, err := redisStore.Subscribe(rec, "incusTesting")
	// if err != nil {
	// 	t.Fatalf("Could not subscribe: %s", err.Error())
	// }
	// defer conn.Quit()

	// go func() {
	// 	time.Sleep(20 * time.Millisecond)
	// 	redisStore.Publish("incusTesting", "TEST")
	// }()

	// <-rec // throwaway subscribe message
	// ms := <-rec

	// if ms[2] != "TEST" {
	// 	t.Fatalf("Subscribe and Publish test failed, got %s", ms[2])
	// }
}

func TestRSave(t *testing.T) {

	// redisStore.Save("TEST")
	// redisStore.Save("TEST1")
	// redisStore.Save("TEST2")
	// redisStore.Save("TEST3")

	// client, err := redisStore.GetConn()
	// if err != nil {
	// 	t.Fatal("Save test failed couldn't get redis connection")
	// }
	// defer client.Quit()

	// arr, _ := client.SMembers(redisStore.clientsKey)

	// if len(arr) != 4 {
	// 	t.Fatal("Save test failed")
	// }
	return
}

func TestRRemove(t *testing.T) {
	// redisStore.Remove("TEST")
	// redisStore.Remove("TEST1")

	// client, err := redisStore.GetConn()
	// if err != nil {
	// 	t.Fatal("Remove test failed couldn't get redis connection")
	// }
	// defer client.Quit()

	// arr, _ := client.SMembers(redisStore.clientsKey)

	// if len(arr) != 2 {
	// 	t.Fatal("Remove test failed")
	// }
	return
}

func TestRClients(t *testing.T) {
	// arr, err := redisStore.Clients()
	// if err != nil {
	// 	t.Fatal("Clients test failed couldn't get redis connection")
	// }

	// if len(arr) != 2 {
	// 	t.Fatal("Clients test failed")
	// }
	return
}

func TestRCount(t *testing.T) {
	// num, err := redisStore.Count()
	// if err != nil {
	// 	t.Fatal("Clients test failed couldn't get redis connection")
	// }

	// if num != 2 {
	// 	t.Fatal("Clients test failed")
	// }
	return
}
