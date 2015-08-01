package main

import (
	"encoding/json"
	"fmt"
	"github.com/Imgur/incus"
	"github.com/garyburd/redigo/redis"
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
var REDISPORT = 6379

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

func pullMessage(c chan []byte, page, user, command string) {
	lpParams := url.Values{}
	lpParams.Set("user", user)
	lpParams.Set("page", page)
	lpParams.Set("command", command)
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

func doredis(command string, args ...interface{}) chan interface{} {
	resultChan := make(chan interface{})

	go func(resultChan chan interface{}) {
		rds, err := redis.Dial("tcp", fmt.Sprintf("%s:%d", REDISHOST, REDISPORT))
		if err != nil {
			log.Fatalf("Failed to connect to redis: %s", err.Error())
		}
		defer rds.Close()
		result, err := rds.Do(command, args...)
		if err != nil {
			log.Fatalf("Failed to execute command: %s", err.Error())
		}

		select {
		case resultChan <- result:
		default:
		}
	}(resultChan)

	return resultChan
}

func TestReceivingMessageFromLongpollViaLongpoll(t *testing.T) {
	msgChan := make(chan []byte)
	go pullMessage(msgChan, "", "foo", "")
	// Give it a little time to set up LP
	time.Sleep(250 * time.Millisecond)

	go sendCommandLP(`{"command":{"command":"message","user":"foo"},"message":{"event":"foobar","data":{},"time":1}}`, "baz")
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
	go pullMessage(msgChan, "", "bar", "")
	// Give it a little time to set up LP
	time.Sleep(250 * time.Millisecond)

	go doredis("PUBLISH", "Incus", `{"command":{"command":"message","user":"bar"},"message":{"event":"foobar","data":{},"time":1}}`)
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
	go pullMessage(msgChan, "", "baz", "")

	// Give it a little time to set up LP
	time.Sleep(250 * time.Millisecond)

	var clientsKilled int
	doredis("CLIENT", "KILL", "SKIPME", "yes")

	t.Logf("Killed %d clients", clientsKilled)

	// Wait for incus to try to reconnect
	time.Sleep(time.Second)

	go doredis("PUBLISH", "Incus", `{"command":{"command":"message","user":"baz"},"message":{"event":"bazbaz","data":{},"time":1}}`)

	select {
	case msg, ok := <-msgChan:
		if !ok {
			t.Fatalf("Channel unexpectedly closed!")
		}
		t.Logf("Incus successfullly reconnected and sent %s!", msg)
	case <-time.After(20 * time.Second):
		t.Fatalf("Timed out waiting for message")
	}
}

func TestMarksPresent(t *testing.T) {
	doredis("DEL", "ClientPresence:foo")

	msgChan := make(chan []byte)
	go pullMessage(msgChan, "/", "foo", `{"command":{"command":"setpresence"},"message":{"presence":true}}`)

	// Give it a little time to set up LP
	time.Sleep(250 * time.Millisecond)

	cardchan := doredis("ZCARD", "ClientPresence:foo")
	card := <-cardchan

	if card.(int64) != 1 {
		t.Fatalf("Expected cardinality of ClientPresence:foo to be 1, instead %d", card)
	}
}

// Sends b.N self-messages via longpoll
func BenchmarkLongpollSending(b *testing.B) {
	for i := 0; i < b.N; i++ {
		sendCommandLP(`{"command":{"command":"message","user":"fooz"},"message":{"event":"foobar","data":{},"time":1}}`, "fooz")
	}
}
