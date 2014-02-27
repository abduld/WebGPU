package server

import (
	"net/http"
	"net/url"
	"strconv"
	"time"
	. "wb/app/config"
	"wb/app/stats"
	. "wb/app/util"
)

type WorkerState struct {
	MachineProblemNumber    int
	MachineProblemId        int64
	DatasetId               int
	ForGrading              bool
	Program                 string
	ProgramFileName         string
	TemporaryDirectory      string
	CompileCommand          string
	BuildFileName           string
	CompileStdout           string
	CompileStderr           string
	RunCommand              string
	RunStdout               string
	RunStderr               string
	RunFailed               bool
	TimeoutError            bool
	TimeoutValue            float64
	CompilationFailed       bool
	Sandboxed               bool
	SandboxKeyword          string
	CompileStartTime        time.Time
	CompileEndTime          time.Time
	RunStartTime            time.Time
	RunEndTime              time.Time
	RequestStartTime        time.Time
	RequestEndTime          time.Time
	SolutionCorrect         bool
	SolutionMessage         string
	UserOutput              string
	ExecutableFileName      string
	MachineProblemDirectory string
	RunId                   string
	OnAllDatasets           bool
	Language                string
	InternalCData           InternalCData
	MachineProblemConfig    *MachineProblemConfig
}

type WorkerInfo struct {
	Address string
	IP      string
	Port    int
}

type WorkerRequest struct {
	MachineProblemNumber int
	MachineProblemId     int64
	DatasetId            int
	RunId                string
	Program              string
	Language             string
	RequestStartTime     time.Time
	ForGrading           bool
}

func registeredWithServer(t time.Time) {
	addr := MasterAddress + "/worker/register"
	stats.TRACE.Println("Registering worker with master at " + addr)
	port := strconv.Itoa(WorkerPort)
	_, err := http.PostForm(addr,
		url.Values{
			"worker.Address": {WorkerAddress},
			"worker.IP":      {WorkerIP},
			"worker.Port":    {port},
		})
	if err != nil {
		stats.TRACE.Println("Failed to register worker with server")
		stats.TRACE.Println(err)
	} else {
		RegisteredWithServer = true
	}
}

func RegisterWithServer() {
	DoEvery(RegisterInterval, registeredWithServer)
}

func InitWorkerServer() {
	if IsWorker {
		MakeUniques()

		RegisteredWithServer = false
		RegisterWithServer()

		DoEvery(SystemMeasureInterval, func(t time.Time) { stats.LogGPUInformation() })
	}
}
