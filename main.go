package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"syscall"
)

var DEBUG bool
var CLIENT_BROAD bool
var store *Storage

func main() {
	if os.Getenv("GOMAXPROCS") == "" {
		runtime.GOMAXPROCS(runtime.NumCPU())
	}

	store = nil
	signals := make(chan os.Signal, 1)

	defer func() {
		if err := recover(); err != nil {
			log.Printf("FATAL: %s", err)
			shutdown()
		}
	}()

	conf := initConfig()
	initLogger(conf)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)
	InstallSignalHandlers(signals)

	store = initStore(&conf)

	CLIENT_BROAD = conf.GetBool("client_broadcasts")
	server := createServer(&conf, store)

	go server.initAppListener()
	go server.initSocketListener()
	go server.initLongPollListener()
	go server.initPingListener()
	go server.sendHeartbeats()

	go listenAndServeTLS(conf)
	listenAndServe(conf)
}

func listenAndServe(conf Configuration) {
	listenAddr := fmt.Sprintf(":%s", conf.Get("listening_port"))
	err := http.ListenAndServe(listenAddr, nil)
	if err != nil {
		log.Fatal(err)
	}
}

func listenAndServeTLS(conf Configuration) {
	if conf.GetBool("tls_enabled") {
		tlsListenAddr := fmt.Sprintf(":%s", conf.Get("tls_port"))
		err := http.ListenAndServeTLS(tlsListenAddr, conf.Get("cert_file"), conf.Get("key_file"), nil)
		if err != nil {
			log.Println(err)
			log.Fatal(err)
		}
	}
}

func InstallSignalHandlers(signals chan os.Signal) {
	go func() {
		sig := <-signals
		log.Printf("%v caught, incus is going down...", sig)
		shutdown()
	}()
}

func initLogger(conf Configuration) {
	DEBUG = false
	if conf.Get("log_level") == "debug" {
		DEBUG = true
	}
}

func shutdown() {
	if store != nil {
		log.Println("clearing redis memory...")
	}

	log.Println("Terminated")
	os.Exit(0)
}
