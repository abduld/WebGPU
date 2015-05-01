package stats

import (
	"time"
	. "wb/app/config"
	. "wb/app/util"
)

func InitStats() {

	InitStatsLogger()

	DoEvery(SystemMeasureInterval, func(t time.Time) {
		LogGPUInformation()
		LogAppInformation()
	})
	ResetStats()
}

func ResetStats() {
	Packets = []Packet{}
}
