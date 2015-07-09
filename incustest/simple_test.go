package incustest

import (
	"encoding/json"
	"github.com/Imgur/incus"
	"github.com/gosexy/redis"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"testing"
	"time"
)

var INCUSHOST = "127.0.0.1:4000"
var REDISHOST = "127.0.0.1"
var REDISPORT = uint(6379)

func TestMain(m *testing.M) {
	// Block until incus is ready
	log.Printf("Waiting for Incus (%s) to be ready.", INCUSHOST)

	timedout := time.After(5 * time.Second)
	ready := false

	for !ready {
		select {
		case <-timedout:
			log.Fatalf("Timed out waiting for Incus to startup")
		default:
			_, err := net.DialTimeout("tcp", INCUSHOST, time.Millisecond)
			if err != nil {
				log.Printf("Failed to connect to Incus (%s): %s", INCUSHOST, err.Error())
			} else {
				log.Printf("Incus is ready!")
				ready = true
			}

			time.Sleep(time.Second)
		}
	}

	os.Exit(m.Run())
}

func pullMessage(c chan []byte, page, user string) {
	lpParams := url.Values{}
	lpParams.Set("user", user)
	lpParams.Set("page", page)
	resp, err := http.PostForm("http://"+INCUSHOST+"/lp", lpParams)
	if err != nil {
		log.Printf("Error POSTing: %s", err.Error())
		close(c)
		return
	}
	defer resp.Body.Close()
	respBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Error reading response: %s", err.Error())
		close(c)
		return
	}
	c <- respBytes
}

func sendCommandLP(command, fromUser string) {
	lpParams := url.Values{}
	lpParams.Set("command", command)
	lpParams.Set("user", fromUser)
	http.PostForm("http://"+INCUSHOST+"/lp", lpParams)
}

func sendCommandRedis(channel, command string) {
	go func() {
		rds := redis.New()
		err := rds.Connect(REDISHOST, REDISPORT)
		if err != nil {
			log.Fatalf("Failed to connect to redis: %s", err.Error())
		}
		_, err = rds.Publish(channel, command)
		if err != nil {
			log.Fatalf("Failed to publish: %s", err.Error())
		}
	}()
}

func TestReceivingMessageFromLongpollViaLongpoll(t *testing.T) {
	msgChan := make(chan []byte)
	go pullMessage(msgChan, "", "userFoo")
	go sendCommandLP(`{"command":{"command":"message","user":"userFoo"},"message":{"event":"foobar","data":{},"time":1}}`, "bazUser")
	select {
	case msgBytes, ok := <-msgChan:
		if !ok {
			t.Fatalf("Channel unexpectedly closed!")
		}

		var msg incus.Message
		err := json.Unmarshal(msgBytes, &msg)
		if err != nil {
			t.Fatalf("Unexpected error unmarshalling %s: %s", msgBytes, err.Error())
		}

		if msg.Event != "foobar" {
			t.Fatalf("Expected event to be 'foobar', instead %s", msg.Event)
		}
		return
	case <-time.After(20 * time.Second):
		t.Fatalf("Timed out waiting for message")
	}
}

func TestReceivingMessageFromLongpollViaRedis(t *testing.T) {
	msgChan := make(chan []byte)
	go pullMessage(msgChan, "", "userFoo")
	go sendCommandRedis("Incus", `{"command":{"command":"message","user":"userFoo"},"message":{"event":"foobar","data":{},"time":1}}`)
	select {
	case msgBytes, ok := <-msgChan:
		if !ok {
			t.Fatalf("Channel unexpectedly closed!")
		}

		var msg incus.Message
		err := json.Unmarshal(msgBytes, &msg)
		if err != nil {
			t.Fatalf("Unexpected error unmarshalling %s: %s", msgBytes, err.Error())
		}

		if msg.Event != "foobar" {
			t.Fatalf("Expected event to be 'foobar', instead %s", msg.Event)
		}
		return
	case <-time.After(20 * time.Second):
		t.Fatalf("Timed out waiting for message")
	}
}

func TestSurvivesRedisDisconnect(t *testing.T) {
	msgChan := make(chan []byte)
	go pullMessage(msgChan, "", "userFoo")
	rds := redis.New()
	err := rds.Connect(REDISHOST, REDISPORT)
	if err != nil {
		t.Fatalf("Failed to connect to redis: %s", err.Error())
	}

	var clientsKilled int
	rds.Command(&clientsKilled, "CLIENT", "KILL", "TYPE", "pubsub")

	t.Logf("Killed %d pubsub clients", clientsKilled)

	// Wait for incus to try to reconnect
	time.Sleep(time.Second)

	go sendCommandRedis("Incus", `{"command":{"command":"message","user":"userFoo"},"message":{"event":"foobar","data":{},"time":1}}`)

	select {
	case _, ok := <-msgChan:
		if !ok {
			t.Fatalf("Channel unexpectedly closed!")
		}
		t.Logf("Incus successfullly reconnected!")
	case <-time.After(20 * time.Second):
		t.Fatalf("Timed out waiting for message")
	}
}

// Sends b.N self-messages via longpoll
func BenchmarkLongpollSending(b *testing.B) {
	for i := 0; i < b.N; i++ {
		sendCommandLP(`{"command":{"command":"message","user":"userFoo"},"message":{"event":"foobar","data":{},"time":1}}`, "userFoo")
	}
}
