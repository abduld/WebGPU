package stats

import (
	"time"
	. "wb/app/config"
	. "wb/app/util"
)

func InitStats() {

	InitStatsLogger()

	if IsMaster && EnableUDP {
		StartUDPServer()
	}

	if EnableRedisStore {
		InitRedisStats()
	}

	if EnableInfluxStore {
		InitInfluxStats()
	}

	DoEvery(SystemMeasureInterval, func(t time.Time) {
		LogCPUInformation()
		LogAppInformation()
	})
}

func ResetStats() {
	Packets = []Packet{}
}
