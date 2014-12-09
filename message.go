package main

import (
	"encoding/json"
	"errors"
	apns "github.com/anachronistic/apns"
	"log"
	"strings"
	"time"
)

type CommandMsg struct {
	Command map[string]string      `json:"command"`
	Message map[string]interface{} `json:"message,omitempty"`
}

type Message struct {
	Event string                 `json:"event"`
	Data  map[string]interface{} `json:"data"`
	Time  int64                  `json:"time"`
}

func (this *CommandMsg) FromSocket(sock *Socket) {
	command, ok := this.Command["command"]
	if !ok {
		return
	}

	if DEBUG {
		log.Printf("Handling socket message of type %s\n", command)
	}

	switch strings.ToLower(command) {
	case "message":
		if !CLIENT_BROAD {
			return
		}

		if sock.Server.Store.StorageType == "redis" {
			this.forwardToRedis(sock.Server)
			return
		}

		this.sendMessage(sock.Server)

	case "setpage":
		page, ok := this.Command["page"]
		if !ok || page == "" {
			return
		}

		if sock.Page != "" {
			sock.Server.Store.UnsetPage(sock) //remove old page if it exists
		}

		sock.Page = page
		sock.Server.Store.SetPage(sock) // set new page
	}
}

func (this *CommandMsg) FromRedis(server *Server) {
	command, ok := this.Command["command"]
	if !ok {
		return
	}

	if DEBUG {
		log.Printf("Handling redis message of type %s\n", command)
	}

	switch strings.ToLower(command) {

	case "message":
		this.sendMessage(server)

	case "pushios":
		this.pushiOS(server)
	}
}

func (this *CommandMsg) formatMessage() (*Message, error) {
	event, e_ok := this.Message["event"].(string)
	data, b_ok := this.Message["data"].(map[string]interface{})

	if !b_ok || !e_ok {
		return nil, errors.New("Could not format message")
	}

	msg := &Message{event, data, time.Now().UTC().Unix()}

	return msg, nil
}

func (this *CommandMsg) sendMessage(server *Server) {
	user, userok := this.Command["user"]
	page, pageok := this.Command["page"]

	if userok {
		this.messageUser(user, page, server)
	} else if pageok {
		this.messagePage(page, server)
	} else {
		this.messageAll(server)
	}
}

func (this *CommandMsg) pushiOS(server *Server) {
	deviceToken, deviceToken_ok := this.Command["device_token"]
	build, _ := this.Command["build"]

	if !deviceToken_ok {
		log.Println("Device token and build not provided!")
		return
	}

	msg, err := this.formatMessage()
	if err != nil {
		log.Println("Could not format message")
		return
	}

	payload := apns.NewPayload()
	payload.Alert = msg.Data["message_text"]
	payload.Sound = server.Config.Get("ios_push_sound")
	payload.Badge = int(msg.Data["badge_count"].(float64))

	pn := apns.NewPushNotification()
	pn.DeviceToken = deviceToken
	pn.AddPayload(payload)
	pn.Set("payload", msg)

	var apns_url string
	var client *apns.Client

	switch build {

	case "store", "enterprise", "beta":

		client = apns.NewClient(apns_url, server.Config.Get("apns_"+build+"_cert"), server.Config.Get("apns_"+build+"_private_key"))

		if build == "store" || build == "enterprise" {
			apns_url = server.Config.Get("apns_production_url")
		} else {
			apns_url = server.Config.Get("apns_sandbox_url")
		}

	default:
		client = apns.NewClient(apns_url, server.Config.Get("apns_store_cert"), server.Config.Get("apns_store_private_key"))
		apns_url = server.Config.Get("apns_production_url")
	}

	resp := client.Send(pn)
	alert, _ := pn.PayloadString()

	if DEBUG {
		log.Printf("Alert: %s\n", alert)
		log.Printf("Success: %s\n", resp.Success)
		log.Printf("Error: %s\n", resp.Error)
	}
}

func (this *CommandMsg) messageUser(UID string, page string, server *Server) {
	msg, err := this.formatMessage()
	if err != nil {
		return
	}

	user, err := server.Store.Client(UID)
	if err != nil {
		return
	}

	for _, sock := range user {
		if page != "" && page != sock.Page {
			continue
		}

		if !sock.isClosed() {
			sock.buff <- msg
		}
	}
}

func (this *CommandMsg) messageAll(server *Server) {
	msg, err := this.formatMessage()
	if err != nil {
		return
	}

	clients := server.Store.Clients()

	for _, user := range clients {
		for _, sock := range user {
			if !sock.isClosed() {
				sock.buff <- msg
			}
		}
	}

	return
}

func (this *CommandMsg) messagePage(page string, server *Server) {
	msg, err := this.formatMessage()
	if err != nil {
		return
	}

	pageMap := server.Store.getPage(page)
	if pageMap == nil {
		return
	}

	for _, sock := range pageMap {
		if !sock.isClosed() {
			sock.buff <- msg
		}
	}

	return
}

func (this *CommandMsg) forwardToRedis(server *Server) {
	msg_str, _ := json.Marshal(this)
	server.Store.redis.Publish(server.Config.Get("redis_message_channel"), string(msg_str)) //pass the message into redis to send message across cluster
}
