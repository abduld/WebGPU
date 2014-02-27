package config

import (
	"path/filepath"

	"github.com/robfig/revel"
)

var (
	DebugMode         bool
	BasePath          string
	MPFileDirectory   string
	ApplicationSecret string
	ApplicationEmail  string
)

func InitConfig() {

	BasePath = revel.BasePath
	MPFileDirectory = filepath.Join(BasePath, "mp")

	InitConfigReader()

	ApplicationSecret, _ = NestedRevelConfig.String("app.secret")
	ApplicationEmail, _ = NestedRevelConfig.String("app.email")

	if debug, found := NestedRevelConfig.Bool("debug"); found {
		DebugMode = debug
	} else {
		DebugMode = false
	}

	InitLoggerConfig()
	InitStatsConfig()
	InitServerConfig()
	InitMailConfig()

	InitUDPConfig()
	InitInfluxConfig()
	InitRedisConfig()

	if IsWorker {
		InitCCConfig()
		InitNVCCConfig()
		InitPGCCConfig()
		InitCPPAMPConfig()

		InitCompileConfig()
		InitWorkerConfig()
	} else {
		InitGeoIPConfig()
		InitCourseraConfig()
		InitDatabaseConfig()
		InitInfluxConfig()
	}

}

func ResetConfig() {
	machineProblemConfigCache = map[int]*MachineProblemConfig{}
	CommonMachineProblemDescription = ""
}
