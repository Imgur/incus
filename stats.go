package main

import (
	"github.com/PagerDuty/godspeed"
)

type RuntimeStats interface {
	LogStartup()

	LogClientCount(int64)

	LogCommand(cmdType string)
	LogPageMessage()
	LogUserMessage()
	LogBroadcastMessage()
	LogReadMessage()
	LogWriteMessage()

	LogWebsocketConnection()
	LogWebsocketDisconnection()

	LogLongpollConnect()
	LogLongpollDisconnect()

	LogAPNSPush()
	LogAPNSError()

	LogGCMPush()
	LogGCMError()
	LogGCMFailure()
}

type DiscardStats struct{}

func (d *DiscardStats) LogStartup()                {}
func (d *DiscardStats) LogClientCount(int64)       {}
func (d *DiscardStats) LogCommand(cmdType string)  {}
func (d *DiscardStats) LogPageMessage()            {}
func (d *DiscardStats) LogUserMessage()            {}
func (d *DiscardStats) LogBroadcastMessage()       {}
func (d *DiscardStats) LogWebsocketConnection()    {}
func (d *DiscardStats) LogWebsocketDisconnection() {}
func (d *DiscardStats) LogReadMessage()            {}
func (d *DiscardStats) LogWriteMessage()           {}
func (d *DiscardStats) LogLongpollConnect()        {}
func (d *DiscardStats) LogLongpollDisconnect()     {}
func (d *DiscardStats) LogAPNSPush()               {}
func (d *DiscardStats) LogGCMPush()                {}
func (d *DiscardStats) LogAPNSError()              {}
func (d *DiscardStats) LogGCMError()               {}
func (d *DiscardStats) LogGCMFailure()             {}

type DatadogStats struct {
	dog *godspeed.Godspeed
}

func NewDatadogStats(datadogHost string) (*DatadogStats, error) {
	gdsp, err := godspeed.New(datadogHost, godspeed.DefaultPort, false)
	if err == nil {
		return &DatadogStats{gdsp}, nil
	} else {
		return nil, err
	}
}

func (d *DatadogStats) LogStartup() {
	d.dog.Incr("incus.startup", nil)
}

func (d *DatadogStats) LogClientCount(clients int64) {
	d.dog.Gauge("incus.client_count", float64(clients), nil)
}

func (d *DatadogStats) LogCommand(cmdType string) {
	d.dog.Incr("incus.command", nil)
	d.dog.Incr("incus.command."+cmdType, nil)
}

func (d *DatadogStats) LogPageMessage() {
	d.dog.Incr("incus.message", nil)
	d.dog.Incr("incus.message.page", nil)
}

func (d *DatadogStats) LogUserMessage() {
	d.dog.Incr("incus.message", nil)
	d.dog.Incr("incus.message.user", nil)
}

func (d *DatadogStats) LogBroadcastMessage() {
	d.dog.Incr("incus.message", nil)
	d.dog.Incr("incus.message.all", nil)
}

func (d *DatadogStats) LogWebsocketConnection() {
	d.dog.Incr("incus.websocket.connect", nil)
}

func (d *DatadogStats) LogWebsocketDisconnection() {
	d.dog.Incr("incus.websocket.disconnect", nil)
}

func (d *DatadogStats) LogReadMessage() {
	d.dog.Incr("incus.read", nil)
}

func (d *DatadogStats) LogWriteMessage() {
	d.dog.Incr("incus.write", nil)
}

func (d *DatadogStats) LogLongpollConnect() {
	d.dog.Incr("incus.longpoll.connect", nil)
}

func (d *DatadogStats) LogLongpollDisconnect() {
	d.dog.Incr("incus.longpoll.disconnect", nil)
}

func (d *DatadogStats) LogAPNSPush() {
	d.dog.Incr("incus.apns.push", nil)
}

func (d *DatadogStats) LogGCMPush() {
	d.dog.Incr("incus.gcm.push", nil)
}

func (d *DatadogStats) LogAPNSError() {
	d.dog.Incr("incus.apns.error", nil)
}

func (d *DatadogStats) LogGCMError() {
	d.dog.Incr("incus.gcm.error", nil)
}

func (d *DatadogStats) LogGCMFailure() {
	d.dog.Incr("incus.gcm.fail", nil)
}
