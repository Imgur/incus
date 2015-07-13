package incus

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
)

type Configuration struct {
	vars map[string]string
}

func NewConfig() Configuration {
	mymap := make(map[string]string)

	ConfigOption(mymap, "CLIENT_BROADCASTS", "true")
	ConfigOption(mymap, "LISTENING_PORT", "4000")
	ConfigOption(mymap, "CONNECTION_TIMEOUT", "0")
	ConfigOption(mymap, "LOG_LEVEL", "debug")
	_, ddEnabled, _ := ConfigOption(mymap, "DATADOG_ENABLED", "false")
	if ddEnabled == "true" {
		ConfigOption(mymap, "DATADOG_HOST", "127.0.0.1")
	}

	_, redisEnabled, _ := ConfigOption(mymap, "REDIS_ENABLED", "false")
	if redisEnabled == "true" {
		ConfigOption(mymap, "REDIS_MESSAGE_CHANNEL", "Incus")
		ConfigOption(mymap, "REDIS_PORT_6379_TCP_PORT", "6379")
		ConfigOption(mymap, "REDIS_PORT_6379_TCP_ADDR", "127.0.0.1")
		ConfigOption(mymap, "REDIS_MESSAGE_QUEUE", "Incus_Queue")
	}

	_, tlsEnabled, _ := ConfigOption(mymap, "TLS_ENABLED", "false")

	if tlsEnabled == "true" {
		ConfigOption(mymap, "TLS_PORT", "443")
		ConfigOption(mymap, "CERT_FILE", "cert.pem")
		ConfigOption(mymap, "KEY_FILE", "key.pem")
	}

	_, apnsEnabled, _ := ConfigOption(mymap, "APNS_ENABLED", "false")

	if apnsEnabled == "true" {
		fileOption(ConfigOption(mymap, "APNS_STORE_CERT", "myapnsappcert.pem"))
		fileOption(ConfigOption(mymap, "APNS_STORE_PRIVATE_KEY", "myapnsappprivatekey.pem"))
		fileOption(ConfigOption(mymap, "APNS_ENTERPRISE_CERT", "myapnsappcert.pem"))
		fileOption(ConfigOption(mymap, "APNS_ENTERPRISE_PRIVATE_KEY", "myapnsappprivatekey.pem"))
		fileOption(ConfigOption(mymap, "APNS_BETA_CERT", "myapnsappcert.pem"))
		fileOption(ConfigOption(mymap, "APNS_BETA_PRIVATE_KEY", "myapnsappprivatekey.pem"))
		fileOption(ConfigOption(mymap, "APNS_DEVELOPMENT_CERT", "myapnsappcert.pem"))
		fileOption(ConfigOption(mymap, "APNS_DEVELOPMENT_PRIVATE_KEY", "myapnsappprivatekey.pem"))

		ConfigOption(mymap, "APNS_PRODUCTION_URL", "gateway.push.apple.com:2195")
		ConfigOption(mymap, "APNS_SANDBOX_URL", "gateway.sandbox.push.apple.com:2195")
		ConfigOption(mymap, "IOS_PUSH_SOUND", "bingbong.aiff")
	}

	_, gcmEnabled, _ := ConfigOption(mymap, "GCM_ENABLED", "false")

	if gcmEnabled == "true" {
		ConfigOption(mymap, "GCM_API_KEY", "foobar")
		ConfigOption(mymap, "ANDROID_ERROR_QUEUE", "Incus_Android_Error_Queue")
	}

	return Configuration{mymap}
}

func ConfigOption(mymap map[string]string, key, default_value string) (string, string, string) {
	envValue := os.Getenv(key)
	var chosenValue string

	if envValue == "" {
		chosenValue = default_value
	} else {
		chosenValue = envValue
	}

	mymap[strings.ToLower(key)] = chosenValue

	return strings.ToLower(key), chosenValue, envValue
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

// Asserts that the chosen value exists on the local file system by panicking if it doesn't
func fileOption(envName, chosenValue, envValue string) {
	if _, err := os.Stat(chosenValue); err != nil {
		panic(fmt.Errorf("Chosen option %s=%s does not exist!", envName, chosenValue))
	}
}
