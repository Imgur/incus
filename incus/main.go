package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/jtaylor32/incus"
	"github.com/spf13/viper"
)

const (
	defaultConfigFilePath = "./"
	configFilePathUsage   = "config file directory (eg. '/etc/incus/'). Config file must be named 'config.yml'."

	gracefulShutdownTimeout = 5
)

var (
	configFilePath string
	store          *incus.Storage
)

// Inserted at compile time by -ldflags "-X main.BUILD foo"
var BUILD string

func init() {
	flag.StringVar(&configFilePath, "conf", defaultConfigFilePath, configFilePathUsage)

	flag.Parse()
}

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

	incus.NewConfig(configFilePath)
	initLogger()
	log.Printf("Incus built on %s", BUILD)

	InstallSignalHandlers()

	var stats incus.RuntimeStats

	if viper.GetBool("datadog_enabled") {
		stats, _ = incus.NewDatadogStats(viper.GetString("datadog_host"))
	} else {
		stats = &incus.DiscardStats{}
	}

	stats.LogStartup()

	store = incus.NewStore(stats)

	incus.CLIENT_BROAD = viper.GetBool("client_broadcasts")
	server := incus.NewServer(store, stats)

	go server.RecordStats(1 * time.Second)
	go server.LogConnectedClientsPeriodically(20 * time.Second)
	go server.ListenFromRedis()
	go server.ListenFromSockets()
	go server.ListenFromLongpoll()
	go server.MonitorLongpollKillswitch()

	go server.ListenForHTTPPings()
	go server.SendHeartbeatsPeriodically(20 * time.Second)

	go listenAndServeTLS()
	listenAndServe()
}

func listenAndServe() {
	listenAddr := fmt.Sprintf(":%s", viper.GetString("listening_port"))
	err := http.ListenAndServe(listenAddr, nil)
	if err != nil {
		log.Fatal(err)
	}
}

func listenAndServeTLS() {
	if viper.GetBool("tls_enabled") {
		tlsListenAddr := fmt.Sprintf(":%s", viper.GetString("tls_port"))
		err := http.ListenAndServeTLS(tlsListenAddr, viper.GetString("cert_file"), viper.GetString("key_file"), nil)
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

func initLogger() {
	incus.DEBUG = false
	if viper.GetString("log_level") == "debug" {
		incus.DEBUG = true
	}
}

func shutdown() {
	log.Println("Terminated")
	os.Exit(0)
}
