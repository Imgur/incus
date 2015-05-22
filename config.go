package main

import (
	"log"
	"os"
	"strconv"
	"strings"
)

type Configuration struct {
	vars map[string]string
}

func initConfig() Configuration {
	mymap := make(map[string]string)

	ConfigOption(mymap, "CLIENT_BROADCASTS", "true")
	ConfigOption(mymap, "LISTENING_PORT", "4000")
	ConfigOption(mymap, "CONNECTION_TIMEOUT", "0")
	ConfigOption(mymap, "LOG_LEVEL", "debug")

	ConfigOption(mymap, "REDIS_MESSAGE_CHANNEL", "Incus")
	redis_port_chosen, redis_port_env := ConfigOption(mymap, "REDIS_PORT_6379_TCP_PORT", "6379")

	if redis_port_env != "" {
		_, redis_host_chosen := ConfigOption(mymap, "REDIS_PORT_6379_TCP_ADDR", "127.0.0.1")

		mymap["redis_host"] = redis_host_chosen
		mymap["redis_port"] = redis_port_chosen
		mymap["redis_enabled"] = "true"
	} else {
		mymap["redis_enabled"] = "false"
	}

	tls_enabled_env, _ := ConfigOption(mymap, "TLS_ENABLED", "false")

	if tls_enabled_env != "" {
		ConfigOption(mymap, "TLS_PORT", "443")
		ConfigOption(mymap, "CERT_FILE", "cert.pem")
		ConfigOption(mymap, "KEY_FILE", "key.pem")
	}

	return Configuration{mymap}
}

func ConfigOption(mymap map[string]string, key, default_value string) (string, string) {
	env_value := os.Getenv(key)
	var chosen_value string

	if env_value == "" {
		chosen_value = default_value
	} else {
		chosen_value = env_value
	}

	mymap[strings.ToLower(key)] = chosen_value

	return chosen_value, env_value
}

func (this *Configuration) Get(name string) string {
	val, ok := this.vars[name]
	if !ok {
		log.Panicf("Config Error: variable '%s' not found", name)
	}

	return val
}

func (this *Configuration) GetInt(name string) int {
	val, ok := this.vars[name]
	if !ok {
		log.Panicf("Config Error: variable '%s' not found", name)
	}

	i, err := strconv.Atoi(val)
	if err != nil {
		log.Panicf("Config Error: '%s' could not be cast as an int", name)
	}

	return i
}

func (this *Configuration) GetBool(name string) bool {
	val, ok := this.vars[name]
	if !ok {
		return false
	}

	return val == "true"
}
