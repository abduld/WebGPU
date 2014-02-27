package config

import "strconv"

var (
	InfluxAddress     string
	InfluxUserName    string
	InfluxPassword    string
	InfluxDB          string
	EnableInfluxStore bool
)

func InitInfluxConfig() {
	conf := NestedRevelConfig

	EnableInfluxStore, _ = conf.Bool("influxdb.enable")
	if EnableInfluxStore {
		InfluxAddress, _ = conf.String("influxdb.addr")
		InfluxUserName, _ = conf.String("influxdb.user")
		InfluxPassword, _ = conf.String("influxdb.password")
		InfluxDB, _ = conf.String("influxdb.db")
		port, _ := conf.Int("influxdb.port")
		InfluxAddress += ":" + strconv.Itoa(port)
	}
}
