package stats

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
	"runtime"
	"strings"
	"time"
	. "wb/app/config"
	. "wb/app/util"

	"bitbucket.org/bertimus9/systemstat"
)

func doLogGPUStatus() {
	if EnableNVML {
		devices, err := gpuIndexedDeviceName()
		if err != nil {
			return
		}
		if fans, err := gpuFanSpeeds(); err == nil {
			for ii, fan := range fans {
				s := fmt.Sprintf("%v", fan)
				Log(ServerRole+"-SystemGPU(", devices[ii]+")", "FanSpeed", s)
			}
		}
		if mems, err := gpuMemoryInformation(); err == nil {
			for ii, mem := range mems {
				if js, err := json.Marshal(mem); err == nil {
					Log(ServerRole+"-SystemGPU(", devices[ii]+")", "Memory", string(js))
				}
			}
		}
	} else {
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
}

func logGPUStatus() {
	var stdout, stderr bytes.Buffer
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

func LogCPUInformation() {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				return
			}
		}()
		memSample := systemstat.GetMemSample()
		cpuSample := systemstat.GetCPUSample()
		uptimeSample := systemstat.GetUptime()

		if js, err := json.Marshal(memSample); err == nil {
			Log(ServerRole+"-System", "Memory", string(js))
		}
		if js, err := json.Marshal(cpuSample); err == nil {
			Log(ServerRole+"-System", "CPU", string(js))
		}
		if js, err := json.Marshal(uptimeSample); err == nil {
			Log(ServerRole+"-System", "UpTime", string(js))
		}
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
