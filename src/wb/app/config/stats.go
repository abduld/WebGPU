package config

import "time"

var (
	EnableNVML            bool
	SystemMeasureInterval time.Duration
)

func InitStatsConfig() {

	conf := NestedRevelConfig

	EnableNVML = true

	t, _ := conf.Int("stats.measure_interval")
	SystemMeasureInterval = time.Duration(t) * time.Second

}
