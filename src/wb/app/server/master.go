package server

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"strings"
	"time"
	"wb/app/models"
	"wb/app/stats"

	"github.com/revel/revel"
)

var (
	Workers map[string]*WorkerInfo = map[string]*WorkerInfo{}
)

func RegisterWorker(w *WorkerInfo) {
	if strings.Contains(w.Address, "%") {
		return
	}
	if _, p := Workers[w.Address]; p == false {
		Workers[w.Address] = w
		revel.TRACE.Println("Added worker...")
		stats.Incr("Master", "Workers")
	}
}

func UnregisterWorker(w *WorkerInfo) {
	delete(Workers, w.Address)
}

func UnregisterWorkerByAddress(addr string) {
	delete(Workers, addr)
}

func chooseWorker() (w *WorkerInfo, err error) {
	var workers []*WorkerInfo
	for _, k := range Workers {
		workers = append(workers, k)
	}
	sz := len(workers)
	if sz == 0 {
		stats.Log("Master", "Error", "Was not able to find a worker.")
		stats.ERROR.Println("Was not able to find a worker.")
		err = errors.New("No workers")
	} else if sz == 1 {
		w = workers[0]
	} else {
		w = workers[rand.Int31n(int32(sz))]
	}
	return
}

func SubmitJob(ws WorkerRequest) (res *WorkerState, err error) {

	worker, err := chooseWorker()
	if err != nil {
		return
	}
	js, err := json.Marshal(ws)
	if err != nil {
		return
	}

	b := bytes.NewBufferString(string(js))

	stats.Incr("Master", "SubmittingJob")

	resp, err := http.Post(worker.Address+"/compile", "text/json", b)
	if err != nil {
		return
	}

	defer resp.Body.Close()

	return
}

func SubmitProgram(mp models.MachineProblem, program string, datasetId int,
	lang string, runId string, forGrading bool) {
	req := WorkerRequest{
		MachineProblemId:     mp.Id,
		MachineProblemNumber: mp.Number,
		DatasetId:            datasetId,
		Program:              program,
		Language:             lang,
		RequestStartTime:     time.Now(),
		RunId:                runId,
		ForGrading:           forGrading,
	}
	SubmitJob(req)
}

// TODO: This should not live here
func CreateAttemptWithStates(states []WorkerState) (attempts []models.Attempt, grade models.Grade, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = errors.New(fmt.Sprint(r))
		}
	}()

	if len(states) == 0 {
		err = errors.New("Invalid attempt")
		return
	}

	firstState := states[0]

	mp, err := models.FindMachineProblem(firstState.MachineProblemId)
	if err != nil {
		return
	}

	prog, err := models.CreateProgram(mp, firstState.Program)
	if err != nil {
		return
	}

	toGrade := false

	for _, ws := range states {

		jsInternalCData, _ := json.Marshal(ws.InternalCData)

		attempt := models.Attempt{
			MachineProblemInstanceId: mp.Id,
			ProgramInstanceId:        prog.Id,
			RunId:                    ws.RunId,
			DatasetId:                ws.DatasetId,
			CompilationFailed:        ws.CompilationFailed,
			CompileStderr:            strings.TrimSpace(ws.CompileStderr),
			CompileStdout:            strings.TrimSpace(ws.CompileStdout),
			RunFailed:                ws.RunFailed,
			RunStdout:                strings.TrimSpace(ws.RunStdout),
			RunStderr:                strings.TrimSpace(ws.RunStderr),
			TimeoutError:             ws.TimeoutError,
			TimeoutValue:             ws.TimeoutValue,
			Sandboxed:                ws.Sandboxed,
			SandboxKeyword:           ws.SandboxKeyword,
			CompileElapsedTime:       ws.CompileEndTime.Sub(ws.CompileStartTime).Nanoseconds(),
			RunElapsedTime:           ws.RunEndTime.Sub(ws.RunStartTime).Nanoseconds(),
			CompileStartTime:         ws.CompileStartTime,
			CompileEndTime:           ws.CompileEndTime,
			RunStartTime:             ws.RunStartTime,
			RunEndTime:               ws.RunEndTime,
			RequestStartTime:         ws.RequestStartTime,
			RequestEndTime:           ws.RequestEndTime,
			SolutionCorrect:          ws.SolutionCorrect,
			SolutionMessage:          ws.SolutionMessage,
			UserOutput:               ws.UserOutput,
			Language:                 ws.Language,
			OnAllDatasets:            ws.OnAllDatasets,
			InternalCData:            string(jsInternalCData),
			Created:                  time.Now(),
			Updated:                  time.Now(),
		}

		if ws.ForGrading {
			attempt.GradedQ = true
			toGrade = true
		}

		err = models.DB.Save(&attempt).Error
		if err != nil {
			revel.TRACE.Println("Failed saving attempt..  ", err)
			stats.Incr("Master", "FailedAttemptStore")
		} else {
			attempts = append(attempts, attempt)
		}
	}

	if toGrade {
		stats.Incr("Master", "UpdatingGrade")
		//revel.TRACE.Println("Updating grade...")
		grade, err = models.UpdateGradeWithAttempts(attempts)
		if err != nil {
			revel.TRACE.Println("Error updating grade  ", err)
		}
	}

	return
}
