package incus

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

var socketIds chan string

func init() {
	socketIds = make(chan string)

	go func() {
		var i = 1
		for {
			i++
			socketIds <- fmt.Sprintf("%v", i)
		}
	}()
}

func newSocket(ws *websocket.Conn, lp http.ResponseWriter, server *Server, UID string) *Socket {
	return &Socket{
		SID:    <-socketIds,
		UID:    UID,
		ws:     ws,
		lp:     lp,
		Server: server,
		buff:   make(chan *Message, 1000),
		done:   make(chan bool),
		closed: false,
		lock:   sync.Mutex{},
	}
}

type Socket struct {
	SID  string // socket ID, randomly generated
	UID  string // User ID, passed in via client
	Page string // Current page, if set.

	// TODO add in group authentication (This will be able to work like people can be assigned to many groups)
	// Group []string // Users can have assigned groups

	ws     *websocket.Conn
	lp     http.ResponseWriter
	Server *Server

	buff   chan *Message
	done   chan bool
	closed bool

	// The purpose of this mutex is to prevent writing to the closed channel buff.
	lock sync.Mutex
}

func (this *Socket) isWebsocket() bool {
	return (this.ws != nil)
}

func (this *Socket) isLongPoll() bool {
	return (this.lp != nil)
}

func (this *Socket) isClosed() bool {
	return this.closed
}

func (this *Socket) Close() error {
	this.lock.Lock()
	defer this.lock.Unlock()

	if !this.closed {
		this.closed = true

		if this.Page != "" {
			this.Server.Store.UnsetPage(this)
			this.Page = ""
		}

		this.Server.Store.redis.MarkInactive(this.UID, this.SID)

		this.Server.Store.Remove(this)
		close(this.done)
	}

	return nil
}

func (this *Socket) Authenticate(UID string) error {

	if this.isWebsocket() {
		var message = new(CommandMsg)
		err := this.ws.ReadJSON(message)

		if DEBUG {
			log.Println(message.Command)
		}
		if err != nil {
			return err
		}

		command := message.Command["command"]
		if strings.ToLower(command) != "authenticate" {
			return errors.New("error: authenticate expected")
		}

		var ok bool
		UID, ok = message.Command["user"]
		if !ok {
			return errors.New("error on authenticate: bad input")
		}
	}

	if UID == "" {
		return errors.New("error on authenticate: bad input")
	}

	if DEBUG {
		log.Printf("saving UID as %s", UID)
	}
	this.UID = UID
	this.Server.Store.Save(this)

	return nil
}

func (this *Socket) listenForMessages() {
	for {

		select {
		case <-this.done:
			return

		default:
			var command = new(CommandMsg)
			err := this.ws.ReadJSON(command)
			if err != nil {
				if DEBUG {
					log.Printf("Error: %s\n", err.Error())
				}

				go this.Close()
				return
			}

			this.Server.Stats.LogReadMessage()

			if DEBUG {
				log.Println(command)
			}
			go command.FromSocket(this)
		}
	}
}

func (this *Socket) listenForWrites() {
	for {
		select {
		case message := <-this.buff:
			if DEBUG {
				log.Println("Sending:", message)
			}

			var err error
			if this.isWebsocket() {
				this.ws.SetWriteDeadline(time.Now().Add(writeWait))
				err = this.ws.WriteJSON(message)
			} else {
				jsonStr, _ := json.Marshal(message)

				_, err = fmt.Fprint(this.lp, string(jsonStr))
			}

			this.Server.Stats.LogWriteMessage()

			if this.isLongPoll() || err != nil {
				if DEBUG && err != nil {
					log.Printf("Error: %s\n", err.Error())
				}

				go this.Close()
				return
			}

		case <-this.done:
			return
		}
	}
}
