package controllers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"wb/app/bigcode"
	. "wb/app/config"
	"wb/app/models"
	"wb/app/routes"
	"wb/app/server"

	"github.com/revel/revel"
)

type BigCodeApplication struct {
	PublicApplication
}

var numProgramsToRequest = -1 // -1 means all

func (c BigCodeApplication) getPrograms(mpNumString string, count int) ([]models.Program, revel.Result, error) {
	var progs []models.Program
	resp, err := http.Get(MasterAddress + routes.MasterApplication.Programs(mpNumString, strconv.Itoa(count), ApplicationSecret))
	if err != nil {
		return progs, c.RenderJson(map[string]interface{}{
			"status": "error",
			"data":   "cannot query master for programs",
			"error":  err,
		}), err
	}
	if err := json.NewDecoder(resp.Body).Decode(&progs); err != nil {
		return progs, c.RenderJson(map[string]interface{}{
			"status": "error",
			"data":   "cannot parse server response for programs",
			"error":  err,
		}), err
	}
	return progs, nil, nil
}

func (c BigCodeApplication) LoadDatabase(mpNumString string) revel.Result {
	if IsMaster {
		for _, worker := range server.Workers {
			resp, _ := http.Get(worker.Address + routes.BigCodeApplication.LoadDatabase(mpNumString))
			defer resp.Body.Close()
		}
		return c.RenderJson(map[string]interface{}{
			"status": "success",
			"data":   "sent request to load big code database to workers",
		})
	} else {
		mpNum, _ := strconv.Atoi(mpNumString)
		db, err := bigcode.LoadOrCreateDatabase(mpNum, nil)
		if err != nil {
			progs, errResponse, err := c.getPrograms(mpNumString, numProgramsToRequest)
			if err != nil {
				return errResponse
			}
			db, err = bigcode.LoadOrCreateDatabase(mpNum, progs)
			return c.RenderJson(map[string]interface{}{
				"status": "failed",
				"data":   "failed to create big code database ",
				"error":  err,
			})
		}

		dbJson, _ := json.MarshalIndent(db, "", "  ")
		return c.RenderText(string(dbJson))
	}
}

func (c BigCodeApplication) DeleteDatabase(mpNumString string) revel.Result {

	return c.Todo()
}

func (c BigCodeApplication) CreateDatabase(mpNumString string) revel.Result {
	if IsMaster {
		for _, worker := range server.Workers {
			resp, _ := http.Get(worker.Address + routes.BigCodeApplication.CreateDatabase(mpNumString))
			defer resp.Body.Close()
		}
		return c.RenderJson(map[string]interface{}{
			"status": "success",
			"data":   "sent request to create big code database to workers",
		})
	} else {
		progs, errResponse, err := c.getPrograms(mpNumString, numProgramsToRequest)
		if err != nil {
			return errResponse
		}
		createBigCodeDatabaseMutex.Lock()
		mpNum, _ := strconv.Atoi(mpNumString)
		db, err := bigcode.CreateDatabase(mpNum, progs)
		createBigCodeDatabaseMutex.Unlock()
		if err != nil {
			return c.RenderJson(map[string]interface{}{
				"status": "failed",
				"data":   "failed to create big code database ",
				"error":  err,
			})
		}

		dbJson, _ := json.MarshalIndent(db, "", "  ")
		return c.RenderText(string(dbJson))
	}
}

func (c BigCodeApplication) ClosestDatabaseEntry(mpNumString string, sig string) revel.Result {
	mpNum, _ := strconv.Atoi(mpNumString)
	entry, err := bigcode.ClosestDatabaseEntry(mpNum, sig)
	if err != nil {
		return c.RenderJson(map[string]interface{}{
			"status": "failed",
			"data":   "failed to find closest match in database",
			"error":  err,
		})
	}

	entryJson, _ := json.MarshalIndent(entry, "", "  ")
	return c.RenderText(string(entryJson))
}

func (c BigCodeApplication) GetSuggestion(mpNumString string) revel.Result {
	if IsMaster {
		return c.RenderJson(map[string]interface{}{
			"status": "failed",
			"data":   "this is a worker only service ",
		})
	} else {
		var prog models.Program

		mpNum, err := strconv.Atoi(mpNumString)
		if err != nil {
			return c.RenderJson(map[string]interface{}{
				"status": "error",
				"data":   "Invalid mp Number",
				"error":  err,
			})
		}
		if err := json.NewDecoder(c.Request.Body).Decode(&prog); err != nil {
			return c.RenderJson(map[string]interface{}{
				"status": "error",
				"data":   "Failed to parse request",
				"error":  err,
			})
		}

		bg, err := bigcode.GetSuggestion(mpNum, prog)
		if err != nil {
			return c.RenderJson(map[string]interface{}{
				"status": "error",
				"data":   "Failed to get suggestion request",
				"error":  err,
			})
		}
		js, _ := json.MarshalIndent(bg, "", "  ")
		return c.RenderText(string(js))
	}
}

func (c BigCodeApplication) Vote(mpNumString, userProgramIdString,
	suggestedProgramIdString, suggestionType, yesOrNo string) revel.Result {
	mpNum, err := strconv.Atoi(mpNumString)
	if err != nil {
		return c.RenderJson(map[string]interface{}{
			"status": "error",
			"data":   "Failed to parse the mp number ",
			"error":  err,
		})
	}
	userProgramId, err := strconv.Atoi(userProgramIdString)
	if err != nil {
		return c.RenderJson(map[string]interface{}{
			"status": "error",
			"data":   "Failed to parse the program id ",
			"error":  err,
		})
	}
	suggestedProgramId, err := strconv.Atoi(suggestedProgramIdString)
	if err != nil {
		return c.RenderJson(map[string]interface{}{
			"status": "error",
			"data":   "Failed to parse the suggested program id ",
			"error":  err,
		})
	}
	if suggestionType != "max" && suggestionType != "min" && suggestionType != "rand" {
		return c.RenderJson(map[string]interface{}{
			"status": "error",
			"data":   "Invalid suggestion type ",
		})
	}
	if yesOrNo != "yes" && yesOrNo != "no" {
		return c.RenderJson(map[string]interface{}{
			"status": "error",
			"data":   "Invalid vote",
		})
	}
	_, err = bigcode.Vote(mpNum, int64(userProgramId), int64(suggestedProgramId), suggestionType, yesOrNo)
	if err != nil {
		return c.RenderJson(map[string]interface{}{
			"status": "error",
			"data":   "Failed to record vote ",
			"error":  err,
		})
	}
	return c.RenderJson(map[string]interface{}{
		"status": "success",
		"data":   "Thank you for voting",
	})
}
