package config

import "strconv"

var (
	EnableUDP  bool
	UDPPort    int
	UDPAddress string
)

func InitUDPConfig() {

	conf := NestedRevelConfig

	EnableUDP, _ = conf.Bool("udp_server.enable")

	if EnableUDP {
		UDPPort, _ = conf.Int("udp_server.port")
		UDPAddress, _ = conf.String("udp_server.ip")
		UDPAddress += ":" + strconv.Itoa(UDPPort)
	}
}
