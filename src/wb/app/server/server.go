package server

import (
	"math/rand"
	. "wb/app/config"
)

func InitServer() {
	if IsWorker {
		InitWorkerServer()
	}
}

func ResetServer() {
	if IsMaster {
		Workers = make(map[string]*WorkerInfo)
	} else {
		RegisteredWithServer = false
	}
}

var Uniq = make(chan int64)

func MakeUniques() {
	go func() {
		for {
			Uniq <- rand.Int63()
		}
	}()
}
