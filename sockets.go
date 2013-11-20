package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"

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
	return &Socket{<-socketIds, UID, "", ws, lp, server, make(chan *Message, 1000), make(chan bool), false}
}

type Socket struct {
	SID  string // socket ID, randomly generated
	UID  string // User ID, passed in via client
	Page string // Current page, if set.

	ws     *websocket.Conn
	lp     http.ResponseWriter
	Server *Server

	buff   chan *Message
	done   chan bool
	closed bool
}

func (this *Socket) isWebsocket() bool {
	return (this.ws != nil)
}

func (this *Socket) isLongPoll() bool {
	return (this.lp != nil)
}

func (this *Socket) Close() error {
	if !this.closed {
		this.closed = true

		if this.Page != "" {
			this.Server.Store.UnsetPage(this)
			this.Page = ""
		}

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
			return errors.New("Error: Authenticate Expected.\n")
		}

		var ok bool
		UID, ok = message.Command["user"]
		if !ok {
			return errors.New("Error on Authenticate: Bad Input.\n")
		}
	}

	if UID == "" {
		return errors.New("Error on Authenticate: Bad Input.\n")
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
				err = this.ws.WriteJSON(message)
			} else {
				json_str, _ := json.Marshal(message)

				_, err = fmt.Fprint(this.lp, string(json_str))
			}

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
