package controllers

import (
	"encoding/json"
	"net/http"
	"time"
	. "wb/app/config"
	"wb/app/server"
	"wb/app/stats"

	"github.com/robfig/revel"
)

type MasterApplication struct {
	PublicApplication
}

func (c MasterApplication) Workers() revel.Result {
	return c.RenderJson(server.Workers)
}

func (c MasterApplication) WorkerRegister() revel.Result {
	worker := new(server.WorkerInfo)
	c.Params.Bind(worker, "worker")
	server.RegisterWorker(worker)
	return c.RenderJson("success")
}

func (c MasterApplication) Attempt() revel.Result {
	w := new(CompileRunResult)

	if err := json.NewDecoder(c.Request.Body).Decode(w); err != nil {
		stats.TRACE.Println("Failed to decode attempt ", err)
		return c.RenderJson(map[string]interface{}{
			"status": "error",
			"data":   "Failed to parse request",
		})
	}

	//revel.TRACE.Println("/here...")

	for _, ws := range w.Result {
		ws.RequestEndTime = time.Now()
		stats.LogTime("Master", "Submission", ws.RequestStartTime, ws.RequestEndTime)
	}

	server.CreateAttemptWithStates(w.Result)

	return c.RenderJson(map[string]interface{}{
		"status": "success",
		"data":   "Attempt logged",
	})
}

func ResetWorkerCache() {
	for _, worker := range server.Workers {
		resp, _ := http.Get(worker.Address + "/reset")
		defer resp.Body.Close()
	}
}
