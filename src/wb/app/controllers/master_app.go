package controllers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"
	. "wb/app/config"
	"wb/app/models"
	"wb/app/server"
	"wb/app/stats"

	"github.com/revel/revel"
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

func (c MasterApplication) Programs(mpNumString, countString string, appSecret string) revel.Result {
	var progs []models.Program
	var mpNum int
	var count int
	var err error
	if appSecret != ApplicationSecret {
		return c.RenderJson(map[string]interface{}{
			"status": "error",
			"data":   "Unauthorized request",
		})
	}
	if mpNum, err = strconv.Atoi(mpNumString); err != nil {
		return c.RenderJson(map[string]interface{}{
			"status": "error",
			"data":   "Was not able to parse the MPNum paramter as an integer",
			"error":  err,
			"mpNum":  mpNumString,
		})
	}
	if count, err = strconv.Atoi(countString); err != nil {
		return c.RenderJson(map[string]interface{}{
			"status": "error",
			"data":   "Was not able to parse the count paramter as an integer",
			"error":  err,
			"count":  countString,
		})
	}
	if count == -1 {
		progs, err = models.AllCorrectProgramsForMP(mpNum)
	} else {
		progs, err = models.CorrectProgramsForMP(mpNum, count)
	}
	if err != nil {
		return c.RenderJson(map[string]interface{}{
			"status": "error",
			"data":   "Was not able to query the database",
			"error":  err,
		})
	}

	return c.RenderJson(progs)

}
