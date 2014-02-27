package controllers

import (
	"bytes"
	"encoding/json"
	"net/http"
	. "wb/app/config"
	"wb/app/server"

	"github.com/robfig/revel"
)

type WorkerApplication struct {
	PublicApplication
}

type CompileRunResult struct {
	Result []server.WorkerState
}

func (c WorkerApplication) Compile() revel.Result {
	w := new(server.WorkerRequest)

	if err := json.NewDecoder(c.Request.Body).Decode(&w); err != nil {
		return c.RenderJson(map[string]interface{}{
			"status": "error",
			"data":   "Failed to parse request",
		})
	}

	resp := CompileRunResult{
		Result: server.CompileAndRun(w),
	}

	if js, err := json.Marshal(resp); err == nil {
		b := bytes.NewBufferString(string(js))
		//revel.TRACE.Println(MasterAddress + "/attempt")
		if resp, err := http.Post(MasterAddress+"/attempt", "text/json", b); err == nil {
			defer resp.Body.Close()
		}
	}

	return c.RenderJson("success")
}

func (c WorkerApplication) GetData(runIdString string) revel.Result {
	return c.Todo()
}
