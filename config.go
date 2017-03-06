package incus

import (
	"fmt"
	"os"

	"github.com/spf13/viper"
)

func NewConfig(configFilePath string) {
	viper.SetConfigName("config")
	viper.AddConfigPath(configFilePath)

	err := viper.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("fatal error config file: %s", err))
	}

	ConfigOption("client_broadcasts", true)
	ConfigOption("listening_port", "4000")
	ConfigOption("connection_timeout", 60000)
	ConfigOption("log_level", "debug")

	ConfigOption("datadog_enabled", false)

	if viper.GetBool("datadog_enabled") {
		ConfigOption("datadog_host", "127.0.0.1")
	}

	ConfigOption("longpoll_killswitch", "longpoll_killswitch")
	ConfigOption("redis_enabled", false)

	if viper.GetBool("redis_enabled") {
		ConfigOption("redis_port_6379_tcp_addr", "127.0.0.1")
		ConfigOption("redis_port_6379_tcp_port", 6379)
		ConfigOption("redis_message_channel", "Incus")
		ConfigOption("redis_message_queue", "Incus_Queue")
		ConfigOption("redis_activity_consumers", 8)
		ConfigOption("redis_connection_pool_size", 20)
	}

	ConfigOption("tls_enabled", false)

	if viper.GetBool("tls_enabled") {
		ConfigOption("tls_port", "443")
		fileOption(ConfigOption("cert_file", "cert.pem"))
		fileOption(ConfigOption("key_file", "key.pem"))
	}

	ConfigOption("apns_enabled", false)

	if viper.GetBool("apns_enabled") {
		fileOption(ConfigOption("apns_store_cert", "myapnsappcert.pem"))
		fileOption(ConfigOption("apns_store_private_key", "myapnsappprivatekey.pem"))

		fileOption(ConfigOption("apns_enterprise_cert", "myapnsappcert.pem"))
		fileOption(ConfigOption("apns_enterprise_private_key", "myapnsappprivatekey.pem"))

		fileOption(ConfigOption("apns_beta_cert", "myapnsappcert.pem"))
		fileOption(ConfigOption("apns_beta_private_key", "myapnsappprivatekey.pem"))

		fileOption(ConfigOption("apns_development_cert", "myapnsappcert.pem"))
		fileOption(ConfigOption("apns_development_private_key", "myapnsappprivatekey.pem"))

		ConfigOption("apns_store_url", "gateway.push.apple.com:2195")
		ConfigOption("apns_enterprise_url", "gateway.push.apple.com:2195")
		ConfigOption("apns_beta_url", "gateway.push.apple.com:2195")
		ConfigOption("apns_development_url", "gateway.sandbox.push.apple.com:2195")

		ConfigOption("apns_production_url", "gateway.push.apple.com:2195")
		ConfigOption("apns_sandbox_url", "gateway.sandbox.push.apple.com:2195")

		ConfigOption("ios_push_sound", "bingbong.aiff")
	}

	ConfigOption("gcm_enabled", false)

	if viper.GetBool("gcm_enabled") {
		ConfigOption("gcm_api_key", "foobar")
		ConfigOption("android_error_queue", "Incus_Android_Error_Queue")
	}
}

func ConfigOption(key string, default_value interface{}) string {
	viper.SetDefault(key, default_value)

	return key
}

// Asserts that the chosen value exists on the local file system by panicking if it doesn't
func fileOption(key string) {
	chosenValue := viper.GetString(key)

	if _, err := os.Stat(chosenValue); err != nil {
		panic(fmt.Errorf("chosen option %s does not exist", chosenValue))
	}
}
