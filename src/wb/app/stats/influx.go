package stats

import (
	. "wb/app/config"

	"github.com/abduld/influxdbc"
)

var (
	InfluxClient influxdbc.InfluxDB
)

func InitInfluxStats() {
	InfluxClient = influxdbc.InfluxDB{InfluxAddress, InfluxDB, InfluxUserName, InfluxPassword}
}
