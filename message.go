package incus

import (
	"encoding/json"
	"errors"
	"log"
	"strings"
	"time"

	"github.com/alexjlockwood/gcm"
	apns "github.com/anachronistic/apns"
	"github.com/spf13/viper"
)

type CommandMsg struct {
	Command map[string]string      `json:"command"`
	Message map[string]interface{} `json:"message,omitempty"`
}

type Message struct {
	Event string                 `json:"event"`
	Data  map[string]interface{} `json:"data"`
	Time  int64                  `json:"time"`
	Url   string                 `json:"internal_url,omitempty"`
	Title string                 `json:"title,omitempty"`
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
	case "setpresence":
		active, ok := this.Message["presence"]

		if !ok {
			if DEBUG {
				log.Printf("Ignoring presence command with no boolean presence")
			}

			return
		}

		switch activeT := active.(type) {
		case bool:
			if activeT {
				sock.Server.Store.redis.MarkActive(sock.UID, sock.SID, time.Now().Unix())
			} else {
				sock.Server.Store.redis.MarkInactive(sock.UID, sock.SID)
			}
		default:
			if DEBUG {
				log.Printf("Ignoring presence command with invalid (non-boolean) type")
			}
		}
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
		if viper.GetBool("apns_enabled") {
			this.pushiOS(server)
		}

	case "pushandroid":
		if viper.GetBool("gcm_enabled") {
			this.pushAndroid(server)
		}

	case "push":
		if strings.ToLower(this.Command["push_type"]) == "ios" {
			this.pushiOS(server)
		}

		if strings.ToLower(this.Command["push_type"]) == "android" {
			this.pushAndroid(server)
		}
	case "pushormessage":

		active, err := server.Store.redis.QueryIsUserActive(this.Command["user"], time.Now().Unix())

		if err == nil {
			if active {
				websocketMessage := &CommandMsg{
					Command: this.Command,
					Message: this.Message["websocket"].(map[string]interface{}),
				}

				websocketMessage.sendMessage(server)
			} else {

				pushData, ok := this.Message["push"].(map[string]interface{})

				if !ok {
					return
				}

				iosMessage, ok := pushData["ios"]

				if ok {
					iosCommand := &CommandMsg{
						Command: this.Command,
						Message: iosMessage.(map[string]interface{}),
					}
					iosCommand.pushiOS(server)
				}

				androidMessage, ok := pushData["android"]
				if ok {
					androidCommand := &CommandMsg{
						Command: this.Command,
						Message: androidMessage.(map[string]interface{}),
					}
					androidCommand.pushAndroid(server)
				}

			}
		} else {
			log.Printf("Error fetching whether %s was active: %s", this.Command["user"], err.Error())
		}
	}
}

func (this *CommandMsg) formatMessage() (*Message, error) {
	event, e_ok := this.Message["event"].(string)
	data, b_ok := this.Message["data"].(map[string]interface{})

	if !b_ok || !e_ok {
		return nil, errors.New("Could not format message")
	}

	msg := &Message{
		Event: event,
		Data:  data,
		Time:  time.Now().UTC().Unix(),
	}

	// hack for bad version of Imgur iOS client
	url, url_ok := data["internal_url"].(string)
	if url_ok {
		msg.Url = url
	}

	title, title_ok := data["title"].(string)
	if title_ok {
		msg.Title = title
	}

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
	payload.Sound = viper.GetString("ios_push_sound")

	// allow message or message_text to trigger Alert
	if _, messageExists := msg.Data["message"]; messageExists {
		payload.Alert = msg.Data["message"]
	}
	if _, textExists := msg.Data["message_text"]; textExists {
		payload.Alert = msg.Data["message_text"]
	}

	badgeAmt, hasBadge := msg.Data["badge_count"]
	if hasBadge {
		payload.Badge = int(badgeAmt.(float64))
	}

	pn := apns.NewPushNotification()
	pn.DeviceToken = deviceToken
	pn.AddPayload(payload)
	pn.Set("payload", msg)

	client := server.GetAPNSClient(build)
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

	sender := server.GetGCMClient()

	server.Stats.LogGCMPush()
	gcmResponse, gcmErr := sender.Send(gcmMessage, 2)
	if gcmErr != nil {
		server.Stats.LogGCMError()
		log.Printf("Error (Android): %s\n", gcmErr)
		return
	}

	if gcmResponse.Failure > 0 {
		server.Stats.LogGCMFailure()
		if !viper.GetBool("redis_enabled") {
			log.Println("Could not push to android_error_queue since redis is not enabled")
			return
		}

		failurePayload := map[string]interface{}{"registration_ids": regIDs, "results": gcmResponse.Results}

		msg_str, _ := json.Marshal(failurePayload)
		server.Store.redis.Push(viper.GetString("android_error_queue"), string(msg_str))
	}
}

func (this *CommandMsg) messageUser(UID string, page string, server *Server) {
	msg, err := this.formatMessage()
	if err != nil {
		if DEBUG {
			log.Printf("Error formatting message: %s", err.Error())
		}
		return
	}

	user, err := server.Store.Client(UID)
	if err != nil {
		if DEBUG {
			log.Printf("Skipping UID %s because %s", UID, err.Error())
		}
		return
	}

	server.Stats.LogUserMessage()

	for _, sock := range user {
		if page != "" && page != sock.Page {
			if DEBUG {
				log.Printf("Skipping given page %s != %s", page, sock.Page)
			}

			continue
		}

		if !sock.isClosed() {
			sock.buff <- msg
		} else {
			if DEBUG {
				log.Printf("Skipping because closed")
			}
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
	server.Store.redis.Publish(viper.GetString("redis_message_channel"), string(msg_str)) //pass the message into redis to send message across cluster
}
