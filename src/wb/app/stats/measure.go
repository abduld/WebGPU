package stats

import (
	"bytes"
	"encoding/json"
	"os/exec"
	"runtime"
	"strings"
	"time"
	. "wb/app/config"
	. "wb/app/util"
)

func doLogGPUStatus() {

	var stdout, stderr bytes.Buffer
	logGPU := func(toLog string) {
		cmd := exec.Command("nvidia-smi ",
			"-q", "-d", toLog, "|", "grep", "Gpu", "|", "cut", "-c35-36")
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr

		err := RunCommandWithTimeout(5*time.Second, cmd)
		if err != nil {
			switch err.(type) {
			case TimeoutError:
				Incr(ServerRole+"-System", "GPUUnResponsive")
				return
			}
		} else {
			Log("System", strings.ToTitle(toLog),
				string(stdout.Bytes())+"::"+string(stderr.Bytes()))
		}
	}
	logGPU("MEMORY")
	logGPU("UTILIZATION")
	logGPU("PERFORMANCE")
	logGPU("ECC")
	logGPU("CLOCK")
	logGPU("TEMPERATURE")

}

func logGPUStatus() {
	var stdout, stderr bytes.Buffer
	if OperatingSystem == "MacOSX" {
		SendMessage(ServerRole+"-GPUHeartBeat", "OSX Does not have nvidia-smi")
		return
	}
	cmd := exec.Command("nvidia-smi", "-q")
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := RunCommandWithTimeout(5*time.Second, cmd)
	if err != nil {
		switch err.(type) {
		case TimeoutError:
			Incr(ServerRole+"-System", "GPUUnResponsive")
			return
		}
		Incr("System", "FailNVIDIASMIExec")
	} else {
		SendMessage(ServerRole+"-GPUHeartBeat", string(stdout.Bytes())+"\n"+string(stderr.Bytes()))
		Log(ServerRole+"-GPUHeartBeat", string(stdout.Bytes())+"\n"+string(stderr.Bytes()))
		Incr(ServerRole+"-System", "GPUResponsive")
		doLogGPUStatus()
	}
}

func LogGPUInformation() {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				return
			}
		}()
		logGPUStatus()
	}()
}

func LogAppInformation() {
	go func() {

		defer func() {
			if r := recover(); r != nil {
				return
			}
		}()

		memStats := runtime.MemStats{}
		runtime.ReadMemStats(&memStats)
		Log(ServerRole+"-SystemApp", "Goroutines", runtime.NumGoroutine())
		if js, err := json.Marshal(memStats); err == nil {
			Log(ServerRole+"-SystemApp", "Memory", string(js))
		}
	}()
}
