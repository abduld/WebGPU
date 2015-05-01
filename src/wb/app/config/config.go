package config

import (
	"path/filepath"

	"github.com/revel/revel"
)

var (
	DebugMode         bool
	CourseraMode      bool
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

	if courseraMode, found := NestedRevelConfig.Bool("mode.coursera"); found {
		CourseraMode = courseraMode
	} else {
		CourseraMode = true
	}

	InitLoggerConfig()
	InitStatsConfig()
	InitServerConfig()
	InitMailConfig()
	InitAdminConfig()
	InitGitHubConfig()

	if IsWorker {
		InitCCConfig()
		InitNVCCConfig()
		InitPGCCConfig()
		InitCPPAMPConfig()

		InitCompileConfig()
		InitWorkerConfig()
		InitBigCodeConfig()
	} else {
		if CourseraMode {
			InitCourseraConfig()
		}
		InitDatabaseConfig()
	}

}

func ResetConfig() {
	machineProblemConfigCache = map[int]*MachineProblemConfig{}
	CommonMachineProblemDescription = ""
}
