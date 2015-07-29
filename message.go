package incus

import (
	"encoding/json"
	"errors"
	"log"
	"strings"
	"time"

	"github.com/alexjlockwood/gcm"
	apns "github.com/anachronistic/apns"
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

	sock.Server.Stats.LogCommand("websocket", strings.ToLower(command))

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

	server.Stats.LogCommand("redis", strings.ToLower(command))

	if DEBUG {
		log.Printf("Handling redis message of type %s\n", command)
	}

	switch strings.ToLower(command) {

	case "message":
		this.sendMessage(server)

	case "pushios":
		if server.Config.GetBool("apns_enabled") {
			this.pushiOS(server)
		}

	case "pushandroid":
		if server.Config.GetBool("gcm_enabled") {
			this.pushAndroid(server)
		}

	case "push":
		if strings.ToLower(this.Command["push_type"]) == "ios" {
			this.pushiOS(server)
		}

		if strings.ToLower(this.Command["push_type"]) == "android" {
			this.pushAndroid(server)
		}
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
	deviceToken, deviceTokenOkay := this.Command["device_token"]
	build, buildOkay := this.Command["build"]

	if !deviceTokenOkay {
		log.Println("Device token not provided!")
		return
	}

	if !buildOkay {
		log.Println("Build type not provided!")
		return
	}

	msg, err := this.formatMessage()
	if err != nil {
		log.Println("Could not format message")
		return
	}

	payload := apns.NewPayload()
	payload.Sound = server.Config.Get("ios_push_sound")

	payload.Alert = msg.Data["message_text"]
	badgeAmt, hasBadge := msg.Data["badge_count"]
	if hasBadge {
		payload.Badge = int(msg.Data["badge_count"].(float64))
	}

	pn := apns.NewPushNotification()
	pn.DeviceToken = deviceToken
	pn.AddPayload(payload)
	pn.Set("payload", msg)

	client := apns.NewClient(server.Config.Get("apns_"+build+"_url"), server.Config.Get("apns_"+build+"_cert"), server.Config.Get("apns_"+build+"_private_key"))
	resp := client.Send(pn)
	alert, _ := pn.PayloadString()
	server.Stats.LogAPNSPush()

	if resp.Error != nil {
		server.Stats.LogAPNSError()
		log.Printf("Alert (iOS): %s\n", alert)
		log.Printf("Error (iOS): %s\n", resp.Error)
	}
}

func (this *CommandMsg) pushAndroid(server *Server) {
	registration_ids, registration_ids_ok := this.Command["registration_ids"]

	if !registration_ids_ok {
		log.Println("Registration ID(s) not provided!")
		return
	}

	msg, err := this.formatMessage()
	if err != nil {
		log.Println("Could not format message")
		return
	}

	data := map[string]interface{}{"event": msg.Event, "data": msg.Data, "time": msg.Time}

	regIDs := strings.Split(registration_ids, ",")
	gcmMessage := gcm.NewMessage(data, regIDs...)

	sender := &gcm.Sender{ApiKey: server.Config.Get("gcm_api_key")}

	server.Stats.LogGCMPush()
	gcmResponse, gcmErr := sender.Send(gcmMessage, 2)
	if gcmErr != nil {
		server.Stats.LogGCMError()
		log.Printf("Error (Android): %s\n", gcmErr)
		return
	}

	if gcmResponse.Failure > 0 {
		server.Stats.LogGCMFailure()
		if !server.Config.GetBool("redis_enabled") {
			log.Println("Could not push to android_error_queue since redis is not enabled")
			return
		}

		failurePayload := map[string]interface{}{"registration_ids": regIDs, "results": gcmResponse.Results}

		msg_str, _ := json.Marshal(failurePayload)
		server.Store.redis.Push(server.Config.Get("android_error_queue"), string(msg_str))
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

	server.Stats.LogUserMessage()

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

	server.Stats.LogBroadcastMessage()
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

	server.Stats.LogPageMessage()

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
