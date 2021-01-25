package mqtt

import (
	flag "github.com/spf13/pflag"
)

const (
	// the bind address on which the MQTT broker listens on
	CfgMQTTBindAddress = "mqtt.bindAddress"

	// the port of the WebSocket MQTT broker
	CfgMQTTWSPort = "mqtt.wsPort"
)

func init() {
	flag.String(CfgMQTTBindAddress, "localhost:1883", "the bind address on which the MQTT broker listens on")
	flag.String(CfgMQTTWSPort, "1888", "port of the WebSocket MQTT broker")
}
