package incus

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

const (
	writeWait = 5 * time.Second
	pongWait  = 1 * time.Second
)

type Server struct {
	ID     string
	Config *Configuration
	Store  *Storage
	Stats  RuntimeStats

	timeout time.Duration
}

func NewServer(conf *Configuration, store *Storage) *Server {
	hash := md5.New()
	io.WriteString(hash, time.Now().String())
	id := string(hash.Sum(nil))

	timeout := time.Duration(conf.GetInt("connection_timeout"))

	var runtimeStats RuntimeStats

	if conf.GetBool("datadog_enabled") {
		runtimeStats, _ = NewDatadogStats(conf.Get("datadog_host"))
		runtimeStats.LogStartup()
	} else {
		runtimeStats = &DiscardStats{}
	}

	return &Server{
		ID:      id,
		Config:  conf,
		Store:   store,
		timeout: timeout,
		Stats:   runtimeStats,
	}
}

func (this *Server) ListenFromSockets() {
	Connect := func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			http.Error(w, "Method not allowed", 405)
			return
		}
		//if r.Header.Get("Origin") != "http://"+r.Host {
		//        http.Error(w, "Origin not allowed", 403)
		//        return
		// }

		ws, err := websocket.Upgrade(w, r, nil, 1024, 1024)
		if _, ok := err.(websocket.HandshakeError); ok {
			http.Error(w, "Not a websocket handshake", 400)
			return
		} else if err != nil {
			log.Println(err)
			return
		}

		defer func() {
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

		if this.timeout <= 0 { // if timeout is 0 then wait forever and return when socket is done.
			<-sock.done
			return
		}

		select {
		case <-time.After(this.timeout * time.Second):
			sock.Close()
			return
		case <-sock.done:
			return
		}
	}

	http.HandleFunc("/socket", Connect)
}

func (this *Server) ListenFromLongpoll() {
	LpConnect := func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			r.Body.Close()
			if DEBUG {
				log.Println("Socket Closed")
			}
		}()

		sock := newSocket(nil, w, this, "")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Cache-Control", "private, no-store, no-cache, must-revalidate, post-check=0, pre-check=0")
		w.Header().Set("Connection", "keep-alive")
		//w.Header().Set("Content-Encoding", "gzip")
		w.WriteHeader(200)

		if DEBUG {
			log.Printf("Long poll connected via \n")
		}

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
			var cmd = new(CommandMsg)
			json.Unmarshal([]byte(command), cmd)

			go cmd.FromSocket(sock)
		}

		go sock.listenForWrites()

		select {
		case <-time.After(30 * time.Second):
			sock.Close()
			return
		case <-sock.done:
			return
		}
	}

	http.HandleFunc("/lp", LpConnect)
}

func (this *Server) ListenFromRedis() {
	if !this.Config.GetBool("redis_enabled") {
		return
	}

	subReciever := make(chan []string, 10000)
	queueReciever := make(chan string, 10000)

	subConsumer, err := this.Store.redis.Subscribe(subReciever, this.Config.Get("redis_message_channel"))
	if err != nil {
		log.Fatal("Couldn't subscribe to redis channel")
	}
	defer subConsumer.Quit()

	err = this.Store.redis.Poll(queueReciever, this.Config.Get("redis_message_queue"))
	if err != nil {
		log.Fatal("Couldn't start polling of redis queue")
	}

	if DEBUG {
		log.Println("LISENING FOR REDIS MESSAGE")
	}

	var subMessage []string
	var pollMessage string
	for {
		var cmd = new(CommandMsg)

		select {
		case subMessage = <-subReciever:
			err = json.Unmarshal([]byte(subMessage[2]), cmd)
		case pollMessage = <-queueReciever:
			err = json.Unmarshal([]byte(pollMessage), cmd)
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
		time.Sleep(period)
	}
}

func (this *Server) LogConnectedClientsPeriodically(period time.Duration) {
	for {
		log.Printf("There are %d connected clients\n", this.Store.memory.clientCount)
		time.Sleep(period)
	}
}
