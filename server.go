package incus

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/alexjlockwood/gcm"
	apns "github.com/anachronistic/apns"
	"github.com/gorilla/websocket"
	"github.com/spf13/viper"
)

const (
	writeWait                         = 5 * time.Second
	pongWait                          = 1 * time.Second
	abnormalCloseControlWriteDeadline = 1 * time.Second
	websocketReadBufferSize           = 1024
	websocketWriteBufferSize          = 1024

	// RFC 6455 Section 7
	closeCodeNormal          = 1000
	closeCodeGoingAway       = 1001
	closeCodeUnexpectedError = 1011
)

var (
	disableLongpoll atomic.Value
)

type GCMClient interface {
	Send(*gcm.Message, int) (*gcm.Response, error)
}

type Server struct {
	ID    string
	Store *Storage
	Stats RuntimeStats

	timeout      time.Duration
	apnsProvider func(string) apns.APNSClient
	gcmProvider  func() GCMClient
}

func NewServer(store *Storage, stats RuntimeStats) *Server {
	hash := md5.New()
	io.WriteString(hash, time.Now().String())
	id := string(hash.Sum(nil))

	timeout := time.Duration(viper.GetInt("connection_timeout"))

	if timeout <= 0 {
		panic(fmt.Errorf("connection_timeout <= 0: %+v", timeout))
	}

	apnsProvider := func(build string) apns.APNSClient {
		return apns.NewClient(viper.GetString("apns_"+build+"_url"), viper.GetString("apns_"+build+"_cert"), viper.GetString("apns_"+build+"_private_key"))
	}

	gcmProvider := func() GCMClient {
		return &gcm.Sender{ApiKey: viper.GetString("gcm_api_key")}
	}

	return &Server{
		ID:           id,
		Store:        store,
		timeout:      timeout,
		Stats:        stats,
		apnsProvider: apnsProvider,
		gcmProvider:  gcmProvider,
	}
}

func (this *Server) ListenFromSockets() {
	Connect := func(w http.ResponseWriter, r *http.Request) {
		writtenCloseMessage := false

		exitSignals := make(chan os.Signal)
		signal.Notify(exitSignals, os.Interrupt, syscall.SIGTERM)
		defer signal.Stop(exitSignals)

		if r.Method != "GET" {
			http.Error(w, "Method not allowed", 405)
			return
		}
		//if r.Header.Get("Origin") != "http://"+r.Host {
		//        http.Error(w, "Origin not allowed", 403)
		//        return
		// }

		ws, err := websocket.Upgrade(w, r, nil, websocketReadBufferSize, websocketWriteBufferSize)
		if _, ok := err.(websocket.HandshakeError); ok {
			http.Error(w, "Not a websocket handshake", 400)
			return
		} else if err != nil {
			log.Println(err)
			return
		}

		defer func() {
			if !writtenCloseMessage {
				closeMessage := websocket.FormatCloseMessage(closeCodeUnexpectedError, "")
				ws.WriteControl(websocket.CloseMessage, closeMessage, time.Now().Add(abnormalCloseControlWriteDeadline))
			}

			ws.Close()
			this.Stats.LogWebsocketDisconnection()
			if DEBUG {
				log.Println("Socket Closed")
			}
		}()

		sock := newSocket(ws, nil, this, "")

		this.Stats.LogWebsocketConnection()
		if DEBUG {
			log.Printf("Socket connected via %s\n", ws.RemoteAddr())
		}
		if err := sock.Authenticate(""); err != nil {
			if DEBUG {
				log.Printf("Error: %s\n", err.Error())
			}
			return
		}

		go sock.listenForMessages()
		go sock.listenForWrites()

		select {
		case <-sock.done:
			writtenCloseMessage = closeWebsocket(closeCodeNormal, ws)
			return
		case <-exitSignals:
			writtenCloseMessage = closeWebsocket(closeCodeGoingAway, ws)
			return
		}
	}

	http.HandleFunc("/socket", Connect)
}

func closeWebsocket(closeCode int, ws *websocket.Conn) bool {
	closeMessage := websocket.FormatCloseMessage(closeCode, "")
	err := ws.WriteControl(websocket.CloseMessage, closeMessage, time.Now().Add(1*time.Second))
	return err == nil
}

func (this *Server) ListenFromLongpoll() {
	LpConnect := func(w http.ResponseWriter, r *http.Request) {
		exitSignals := make(chan os.Signal)
		signal.Notify(exitSignals, os.Interrupt, syscall.SIGTERM)
		defer signal.Stop(exitSignals)

		defer func() {
			this.Stats.LogLongpollDisconnect()
			r.Body.Close()
			if DEBUG {
				log.Println("Socket Closed")
			}
		}()

		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Cache-Control", "private, no-store, no-cache, must-revalidate, post-check=0, pre-check=0")
		w.Header().Set("Connection", "keep-alive")
		//w.Header().Set("Content-Encoding", "gzip")

		longpollIsDisabled := disableLongpoll.Load()
		if longpollIsDisabled != nil && longpollIsDisabled.(bool) == true {
			w.Header().Set("Connection", "close")
			w.WriteHeader(503)
			return
		}

		sock := newSocket(nil, w, this, "")

		if DEBUG {
			log.Printf("Long poll connected via \n")
		}

		this.Stats.LogLongpollConnect()

		if err := sock.Authenticate(r.FormValue("user")); err != nil {
			if DEBUG {
				log.Printf("Error: %s\n", err.Error())
			}
			return
		}

		page := r.FormValue("page")
		if page != "" {
			if sock.Page != "" {
				this.Store.UnsetPage(sock) //remove old page if it exists
			}

			sock.Page = page
			this.Store.SetPage(sock)
		}

		command := r.FormValue("command")
		if command != "" {
			this.Stats.LogReadMessage()

			var cmd = new(CommandMsg)
			json.Unmarshal([]byte(command), cmd)

			go cmd.FromSocket(sock)
		}

		go sock.listenForWrites()

		select {
		case <-exitSignals:
			sock.Close()
			w.WriteHeader(503)
			return
		case <-time.After(this.timeout * time.Second):
			sock.Close()
			w.WriteHeader(204)
			return
		case <-sock.done:
			// No need to write 200, as it's already written implicitly.
			return
		}

	}

	http.HandleFunc("/lp", LpConnect)
}

func (this *Server) ListenFromRedis() {
	if !viper.GetBool("redis_enabled") {
		return
	}

	subReciever := make(chan []byte, 10000)
	queueReciever := make(chan []byte, 10000)

	_, err := this.Store.redis.Subscribe(subReciever, viper.GetString("redis_message_channel"))
	if err != nil {
		log.Fatal("Couldn't subscribe to redis channel")
	}

	err = this.Store.redis.Poll(queueReciever, viper.GetString("redis_message_queue"))
	if err != nil {
		log.Fatal("Couldn't start polling of redis queue")
	}

	if DEBUG {
		log.Println("LISTENING FOR REDIS MESSAGE")
	}

	var subMessage []byte
	var pollMessage []byte
	for {
		var cmd = new(CommandMsg)

		select {
		case subMessage = <-subReciever:
			err = json.Unmarshal(subMessage, cmd)
		case pollMessage = <-queueReciever:
			err = json.Unmarshal(pollMessage, cmd)
		}

		if err != nil {
			log.Printf("Error decoding JSON: %s", err.Error())
			this.Stats.LogInvalidJSON()
		} else {
			go cmd.FromRedis(this)
		}
	}
}

func (this *Server) ListenForHTTPPings() {
	pingHandler := func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "OK")
	}

	http.HandleFunc("/ping", pingHandler)
}

func (this *Server) SendHeartbeatsPeriodically(period time.Duration) {

	for {
		time.Sleep(period)

		clients := this.Store.Clients()

		for _, user := range clients {
			for _, sock := range user {

				if sock.isWebsocket() {
					if !sock.isClosed() {
						sock.ws.WriteControl(websocket.PingMessage, []byte{}, time.Now().Add(pongWait))
					}
				}

			}
		}
	}
}

func (this *Server) RecordStats(period time.Duration) {
	for {
		this.Stats.LogClientCount(this.Store.memory.clientCount)
		this.Stats.LogGoroutines(runtime.NumGoroutine())
		time.Sleep(period)
	}
}

func (this *Server) LogConnectedClientsPeriodically(period time.Duration) {
	for {
		log.Printf("There are %d connected clients\n", this.Store.memory.clientCount)
		time.Sleep(period)
	}
}

func (this *Server) GetAPNSClient(build string) apns.APNSClient {
	return this.apnsProvider(build)
}

func (this *Server) GetGCMClient() GCMClient {
	return this.gcmProvider()
}

func (this *Server) MonitorLongpollKillswitch() {
	if !viper.GetBool("redis_enabled") {
		return
	}

	for {
		longpollSwitchedOff, err := this.Store.redis.GetIsLongpollKillswitchActive()

		if err == nil {
			disableLongpoll.Store(longpollSwitchedOff)
		}

		time.Sleep(5 * time.Second)
	}
}
