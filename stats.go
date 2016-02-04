package incus

import (
	"github.com/PagerDuty/godspeed"
	"net"
)

type RuntimeStats interface {
	LogStartup()

	LogClientCount(int64)
	LogGoroutines(int)

	LogCommand(from, cmdType string)
	LogPageMessage()
	LogUserMessage()
	LogBroadcastMessage()
	LogReadMessage()
	LogWriteMessage()
	LogInvalidJSON()

	LogWebsocketConnection()
	LogWebsocketDisconnection()

	LogLongpollConnect()
	LogLongpollDisconnect()

	LogAPNSPush()
	LogAPNSError()

	LogGCMPush()
	LogGCMError()
	LogGCMFailure()

	LogPendingRedisActivityCommandsListLength(int)
}

type DiscardStats struct{}

func (d *DiscardStats) LogStartup()                                   {}
func (d *DiscardStats) LogClientCount(int64)                          {}
func (d *DiscardStats) LogGoroutines(int)                          	  {}
func (d *DiscardStats) LogCommand(from, cmdType string)               {}
func (d *DiscardStats) LogPageMessage()                               {}
func (d *DiscardStats) LogUserMessage()                               {}
func (d *DiscardStats) LogBroadcastMessage()                          {}
func (d *DiscardStats) LogWebsocketConnection()                       {}
func (d *DiscardStats) LogWebsocketDisconnection()                    {}
func (d *DiscardStats) LogReadMessage()                               {}
func (d *DiscardStats) LogWriteMessage()                              {}
func (d *DiscardStats) LogLongpollConnect()                           {}
func (d *DiscardStats) LogLongpollDisconnect()                        {}
func (d *DiscardStats) LogAPNSPush()                                  {}
func (d *DiscardStats) LogGCMPush()                                   {}
func (d *DiscardStats) LogAPNSError()                                 {}
func (d *DiscardStats) LogGCMError()                                  {}
func (d *DiscardStats) LogGCMFailure()                                {}
func (d *DiscardStats) LogInvalidJSON()                               {}
func (d *DiscardStats) LogPendingRedisActivityCommandsListLength(int) {}

type DatadogStats struct {
	dog *godspeed.Godspeed
}

func NewDatadogStats(datadogHost string) (*DatadogStats, error) {
	var ip net.IP = nil
	var err error = nil

	// Assume datadogHost is an IP and try to parse it
	ip = net.ParseIP(datadogHost)

	// Parsing failed
	if ip == nil {
		ips, _ := net.LookupIP(datadogHost)

		if len(ips) > 0 {
			ip = ips[0]
		}
	}

	if ip != nil {
		gdsp, err := godspeed.New(ip.String(), godspeed.DefaultPort, false)
		if err == nil {
			return &DatadogStats{gdsp}, nil
		}
	}

	return nil, err
}

func (d *DatadogStats) LogStartup() {
	d.dog.Incr("incus.startup", nil)
}

func (d *DatadogStats) LogClientCount(clients int64) {
	d.dog.Gauge("incus.client_count", float64(clients), nil)
}

func (d *DatadogStats) LogGoroutines(goroutines int) {
	d.dog.Gauge("incus.goroutines", float64(goroutines), nil)
}

func (d *DatadogStats) LogCommand(from, cmdType string) {
	d.dog.Incr("incus.command", nil)
	d.dog.Incr("incus.command."+from, nil)
	d.dog.Incr("incus.command."+cmdType, nil)
	d.dog.Incr("incus.command."+from+"."+cmdType, nil)
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

func (d *DatadogStats) LogInvalidJSON() {
	d.dog.Incr("incus.jsonerror", nil)
}

func (d *DatadogStats) LogPendingRedisActivityCommandsListLength(length int) {
	d.dog.Gauge("incus.pendingactivityredislen", float64(length), nil)
}
