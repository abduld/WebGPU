package config

import (
	"strconv"
	"strings"

	"github.com/revel/revel"
)

var (
	IsWorker      bool
	IsMaster      bool
	MasterAddress string
	MasterPort    int
	MasterIP      string
	WorkerAddress string
	WorkerPort    int
	WorkerIP      string
	ServerRole    string
	IP            string
	Port          int
	Address       string
	RunMode       string
)

func mirrorSection(section string) string {
	s := strings.Split(section, ".")
	if len(s) == 0 {
		return section
	} else if s[0] == "master" {
		s[0] = "worker"
		return strings.Join(s, ".")
	} else if s[0] == "worker" {
		s[0] = "master"
		return strings.Join(s, ".")
	} else {
		return section
	}
}

func InitServerConfig() {

	conf := NestedRevelConfig

	RunMode := revel.RunMode

	if parentSection(revel.RunMode) == "worker" {
		IsWorker = true
		ServerRole = "Worker"

		WorkerIP, _ = conf.String("worker.ip")
		WorkerPort, _ = conf.Int("worker.port")
		WorkerAddress = "http://" + WorkerIP + ":" + strconv.Itoa(WorkerPort)

		mirror := mirrorSection(RunMode)
		MasterIP, _ = conf.StringInSection("master.ip", mirror)
		MasterPort, _ = conf.IntInSection("master.port", mirror)

	} else {
		IsWorker = false
		ServerRole = "Master"

		MasterIP, _ = conf.String("master.ip")
		MasterPort, _ = conf.Int("master.port")
	}

	IsMaster = !IsWorker

	MasterAddress = "http://" + MasterIP + ":" + strconv.Itoa(MasterPort)

	if IsWorker {
		Address = WorkerAddress
		Port = WorkerPort
	} else {
		Address = MasterAddress
		Port = MasterPort
	}

	revel.TRACE.Println("Address = ", Address)
	revel.TRACE.Println("Port = ", Port)
	revel.TRACE.Println("MasterAddress = ", MasterAddress)
	revel.TRACE.Println("IsWorker = ", IsWorker)
	revel.TRACE.Println("IsMaster = ", IsMaster)

}
