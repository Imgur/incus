package main

import (
	"fmt"
	"github.com/Imgur/incus"
	"log"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"
)

var store *incus.Storage

func main() {
	if os.Getenv("GOMAXPROCS") == "" {
		runtime.GOMAXPROCS(runtime.NumCPU())
	}

	store = nil
	signals := make(chan os.Signal, 1)

	defer func() {
		if err := recover(); err != nil {
			log.Printf("Caught panic in the main thread: %s", err)
			shutdown()
		}
	}()

	conf := incus.NewConfig()
	initLogger(conf)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)
	InstallSignalHandlers(signals)

	store = incus.NewStore(&conf)

	incus.CLIENT_BROAD = conf.GetBool("client_broadcasts")
	server := incus.NewServer(&conf, store)

	go server.RecordStats(1 * time.Second)
	go server.LogConnectedClientsPeriodically(20 * time.Second)
	go server.ListenFromRedis()
	go server.ListenFromSockets()
	go server.ListenFromLongpoll()
	go server.ListenForHTTPPings()
	go server.SendHeartbeatsPeriodically(20 * time.Second)

	go listenAndServeTLS(conf)
	listenAndServe(conf)
}

func listenAndServe(conf incus.Configuration) {
	listenAddr := fmt.Sprintf(":%s", conf.Get("listening_port"))
	err := http.ListenAndServe(listenAddr, nil)
	if err != nil {
		log.Fatal(err)
	}
}

func listenAndServeTLS(conf incus.Configuration) {
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

func initLogger(conf incus.Configuration) {
	incus.DEBUG = false
	if conf.Get("log_level") == "debug" {
		incus.DEBUG = true
	}
}

func shutdown() {
	log.Println("Terminated")
	os.Exit(0)
}
