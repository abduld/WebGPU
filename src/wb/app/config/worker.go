package config

import (
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/robfig/revel"
)

var (
	OperatingSystem           string
	SystemId                  string
	SystemTemporaryDirectory  string
	CompileTimeout            time.Duration
	RuntimeTimeout            time.Duration
	RegisteredWithServer      bool
	RegisterInterval          time.Duration
	SystemExecutableExtension string
)

func findOperatingSystem() {
	switch runtime.GOOS {
	case "windows":
		OperatingSystem = "Windows"
	case "linux":
		OperatingSystem = "Linux"
	case "darwin":
		OperatingSystem = "MacOSX"
	default:
		OperatingSystem = "Unknown"
	}
}

func findSystemId() {
	SystemId = OperatingSystem
	if runtime.GOARCH == "amd64" {
		SystemId += "-x86_64"
	}
}

func findSystemTemporaryDirectory() {
	if dir, found := NestedRevelConfig.String("temporary_directory"); found {
		SystemTemporaryDirectory = dir
	} else {
		if res, err := filepath.EvalSymlinks(os.TempDir()); err == nil {
			SystemTemporaryDirectory = res
		} else {
			revel.ERROR.Println("Cannot find TemporaryDirectory")
			panic("Cannot find TemporaryDirectory")
		}
	}
}

func findSystemExecutableExtension() {
	if OperatingSystem == "Windows" {
		SystemExecutableExtension = ".exe"
	} else {
		SystemExecutableExtension = ""
	}
}

func InitWorkerConfig() {

	t, _ := NestedRevelConfig.Int("worker.register_interval")
	RegisterInterval = time.Duration(t) * time.Second

	t, _ = NestedRevelConfig.Int("worker.compile_timeout")
	CompileTimeout = time.Duration(t) * time.Second

	t, _ = NestedRevelConfig.Int("worker.run_timeout")
	RuntimeTimeout = time.Duration(t) * time.Second

	findOperatingSystem()
	findSystemId()
	findSystemTemporaryDirectory()
	findSystemExecutableExtension()

	revel.TRACE.Println("(Worker) OperatingSystem = ", OperatingSystem)
	revel.TRACE.Println("(Worker) SystemId = ", SystemId)
	revel.TRACE.Println("(Worker) NVCCCompilerLocation = ", NVCCCompilerLocation)
	revel.TRACE.Println("(Worker) CUDAToolkitDirectory = ", CUDAToolkitDirectory)
	revel.TRACE.Println("(Worker) SystemTemporaryDirectory = ", SystemTemporaryDirectory)
}
