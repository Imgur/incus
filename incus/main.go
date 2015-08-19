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

const gracefulShutdownTimeout = 5

var store *incus.Storage

// Inserted at compile time by -ldflags "-X main.BUILD foo"
var BUILD string

func main() {
	if os.Getenv("GOMAXPROCS") == "" {
		runtime.GOMAXPROCS(runtime.NumCPU())
	}

	store = nil

	defer func() {
		if err := recover(); err != nil {
			log.Printf("Caught panic in the main thread: %s", err)
			shutdown()
		}
	}()

	conf := incus.NewConfig()
	initLogger(conf)
	log.Printf("Incus built on %s", BUILD)

	InstallSignalHandlers()

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

func InstallSignalHandlers() {
	go func() {
		signals := make(chan os.Signal, 1)
		signal.Notify(signals, os.Interrupt, syscall.SIGTERM)
		sig := <-signals
		log.Printf("%v caught, incus is going down...", sig)
		log.Printf("Waiting %d seconds for goroutines to shut down...", gracefulShutdownTimeout)

		select {
		case <-time.After(gracefulShutdownTimeout * time.Second):
			shutdown()
		case sig := <-signals:
			log.Printf("%v caught again. Exiting immediately...", sig)
			shutdown()
		}
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
