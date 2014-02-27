package controllers

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
	. "wb/app/config"
	"wb/app/models"
	"wb/app/routes"
	"wb/app/server"
	"wb/app/stats"

	"github.com/nu7hatch/gouuid"
	"github.com/robfig/revel"
)

type SecuredApplication struct {
	PublicApplication
}

func CheckUser(c0 *revel.Controller) revel.Result {
	c := PublicApplication{c0}
	stats.Log("App", "EndRequest", c.Action)
	if user := c.connected(); user.Id == 0 {
		c.Flash.Error("Please login first")
		return c.Redirect(routes.PublicApplication.Index())
	}
	return nil
}

func CheckWorker(c *revel.Controller) revel.Result {
	if WorkerIsAlive() {
		c.RenderArgs["worker_alive"] = true
	} else {
		c.RenderArgs["worker_alive"] = false
	}
	return nil
}

func (c SecuredApplication) Logout() revel.Result {
	for k := range c.Session {
		delete(c.Session, k)
	}
	stats.Incr("App", "Logout")
	return c.Redirect(routes.PublicApplication.Index())
}

func (c SecuredApplication) UserProfile() revel.Result {
	if user := c.connected(); user.Id != 0 {
		return c.RenderJson(user)
	}
	return c.RenderJson(map[string]interface{}{
		"status": "error",
		"result": "Not logged in",
	})
}

func validMpNumber(s string) bool {
	_, err := strconv.Atoi(s)
	return err == nil
}

func (c SecuredApplication) MachineProblemDatasetDownload(
	mpNumString string, name string) revel.Result {

	if !validMpNumber(mpNumString) {
		return c.Redirect("/404")
	}

	if name != "dataset.zip" && name != "dataset.tar.gz" {
		return c.Redirect("/404")
	}

	path := filepath.Join(MPFileDirectory, mpNumString, name)

	if f, err := os.Open(path); err == nil {
		stats.Incr("App", "DatasetDownload")
		return c.RenderFile(f, "attachement")
	}
	return c.Redirect("/404")
}

func (c SecuredApplication) MachineProblems() revel.Result {
	return c.Render()
}

func machineProblemQuestions(user models.User, mp models.MachineProblem) ([]models.QuestionItem, error) {
	questions, err := models.FindOrCreateQuestionsByMachineProblem(mp)
	if err != nil {
		return nil, err
	}

	questionItems, err := models.FindQuestionItems(mp, questions)
	if err != nil {
		stats.TRACE.Println("Cannot get question items.")
		return nil, err
	}

	return questionItems, nil
}

func revertMachineProblemProgram(user models.User, mp models.MachineProblem) (models.Program, error) {
	conf, err := ReadMachineProblemConfig(mp.Number)
	if err != nil {
		return models.Program{}, err
	}
	templateProgram := conf.CodeTemplate
	return models.CreateProgram(mp, templateProgram)
}

func machineProblemProgram(user models.User, mp models.MachineProblem) (models.Program, error) {
	if prog, err := models.RecentProgram(mp); err == nil && prog.Text != "" {
		return prog, err
	}
	return revertMachineProblemProgram(user, mp)
}

func (c SecuredApplication) MachineProblem(mpNumString string) revel.Result {

	mpNum, err := strconv.Atoi(mpNumString)
	if err != nil {
		c.Flash.Error("Invalid machine problem")
		return c.Render(routes.PublicApplication.Index())
	}

	user := c.connected()

	conf, _ := ReadMachineProblemConfig(mpNum)

	if MachineProblemCodingDeadlineExpiredQ(conf) {
		c.RenderArgs["coding_expired"] = true
	}
	if MachineProblemPeerReviewDeadlineExpiredQ(conf) {
		c.RenderArgs["peer_review_expired"] = true
	}

	var mp models.MachineProblem
	if mp, err = models.FindOrCreateMachineProblemByUser(user, mpNum); err != nil {
		c.Flash.Error("Cannot create machine problem")
		return c.Render(routes.PublicApplication.Index())
	}

	c.RenderArgs["title"] = conf.Name
	c.RenderArgs["mp_num"] = mpNumString
	c.RenderArgs["mp_description"] = conf.Description

	c.RenderArgs["mp_config"] = conf

	if prog, err := machineProblemProgram(user, mp); err == nil {
		c.RenderArgs["mp_program"] = prog.Text
	}

	if questions, err := machineProblemQuestions(user, mp); err == nil {
		c.RenderArgs["mp_questions"] = questions
	}

	if history, err := recentPrograms(user, mp, 100); err == nil {
		c.RenderArgs["mp_program_history"] = history
	}

	c.RenderArgs["mp_attempts"] = getAttemptsSummary(user, mp)

	stats.Incr("App", "MachineProblemViews")

	return c.Render()
}

type attemptSummary struct {
	Id            int64     `json:"id"`
	Snippet       string    `json:"snippet"`
	Status        bool      `json:"status"`
	StatusMessage string    `json:"status_message"`
	DatasetId     int       `json:"dataset_id"`
	AlertTag      string    `json:"alert"`
	Created       time.Time `json:"created"`
}

func getAttemptsSummary(user models.User, mp models.MachineProblem) []attemptSummary {
	attempts, err := models.FindAttemptsByMachineProblem(mp)
	if err != nil {
		return nil
	} else {
		attemptListSummary := make([]attemptSummary, len(attempts))
		for i, attempt := range attempts {
			progId := attempt.ProgramInstanceId
			prog, _ := models.FindProgram(progId)
			attemptListSummary[i] = attemptSummary{
				Id:            attempts[i].Id,
				Snippet:       summarizeProgram(prog.Text),
				Status:        attemptFailed(attempt),
				StatusMessage: attemptFailedMessage(attempt),
				AlertTag:      attemptAlertTag(attempt),
				DatasetId:     attempt.DatasetId,
				Created:       attempt.Created,
			}
		}
		return attemptListSummary
	}
}

func attemptFailed(attempt models.Attempt) bool {
	if attempt.CompilationFailed ||
		attempt.Sandboxed ||
		attempt.RunFailed ||
		attempt.TimeoutError ||
		attempt.SolutionCorrect == false {
		return true
	} else {
		return false
	}
}

func attemptFailedReason(attempt models.Attempt) string {
	stats.TRACE.Println(attempt.TimeoutError)
	if attempt.TimeoutError {
		return "Timeout"
	} else if attempt.CompilationFailed {
		return "Compilation Failed."
	} else if attempt.Sandboxed {
		return "Sandboxed"
	} else if attempt.RunFailed {
		return "Program Run Failed."
	} else if attempt.SolutionCorrect == false && attempt.DatasetId >= 0 {
		return "Incorrect Solution."
		return "Solution is not correct for Dataset " + strconv.Itoa(attempt.DatasetId) + "."
	} else if attempt.SolutionCorrect == false {
		return "Incorrect Solution."
	} else {
		return "Solution is correct."
	}
}

func attemptFailedMessage(attempt models.Attempt) (str string) {
	if attempt.TimeoutError {
		timeoutStr := strconv.FormatFloat(attempt.TimeoutValue, 'f', 2, 64)
		return "Program terminated because it took more than `" + timeoutStr + "` seconds to run."
	} else if attempt.CompilationFailed {
		str = "Failed to compile program.\n"
		if attempt.CompileStderr != "" {
			str += "Compiler Stderr: " + attempt.CompileStderr
			str += " . \n"
		}
		if attempt.CompileStdout != "" {
			str += "Compiler Stdout: " + attempt.CompileStdout
		}
	} else if attempt.Sandboxed {
		return "Failed to run program because `" + attempt.SandboxKeyword + "` is a sandboxed keyword."
	} else if attempt.RunFailed {
		str = "Failed to run program.\n"
		if attempt.RunStderr != "" {
			str += "Run Stderr: " + attempt.RunStderr
			str += " . \n"
		}
		if attempt.RunStdout != "" {
			str += "Run Stdout: " + attempt.RunStdout
		}
	} else if attempt.SolutionCorrect == false {
		str = "Solution is not correct.\n"
		if attempt.SolutionMessage != "" {
			str += attempt.SolutionMessage
		}
	} else {
		str = "Solution is correct."
	}
	return
}

func attemptAlertTag(attempt models.Attempt) string {
	if attemptFailed(attempt) {
		return "alert alert-danger"
	} else {
		return ""
	}
}

func generateRunId(user models.User) string {
	u, err := uuid.NewV4()
	if err != nil {
		stats.ERROR.Println("Cannot generate UUID")
		return "CannotGenerateRunId"
	}
	return fmt.Sprint(user.UserName, "-", user.Id, "-", u.String())
}

func (c SecuredApplication) SubmitProgram(mpNumString string) (res revel.Result) {
	var program string
	var datasetId int

	mpNum, err := strconv.Atoi(mpNumString)
	if err != nil {
		c.Flash.Error("Invalid machine problem")
		return c.Render(routes.PublicApplication.Index())
	}

	user := c.connected()

	c.Params.Bind(&program, "program")
	c.Params.Bind(&datasetId, "datasetId")

	mp, err := models.FindOrCreateMachineProblemByUser(user, mpNum)
	if err != nil {
		return c.RenderJson(map[string]interface{}{
			"status": "error",
			"title":  "Error finding MP record",
			"data":   "System was not able to save program.",
		})
	}

	if lastAttempt, err := models.FindLastAttemptByMachineProblem(mp); err == nil && time.Since(lastAttempt.Created).Seconds() < 5 {
		return c.RenderJson(map[string]interface{}{
			"status": "error",
			"title":  "Attempt limiter",
			"data":   "Too many attempts. Please wait 10 seconds between attempts.",
		})
	}

	if _, err := models.CreateProgram(mp, program); err != nil {
		return c.RenderJson(map[string]interface{}{
			"status": "error",
			"title":  "Error Saving Program",
			"data":   "System was not able to save program.",
		})
	}

	runId := generateRunId(user)

	conf, _ := ReadMachineProblemConfig(mpNum)

	stats.Incr("App", "ProgramSubmission")

	server.SubmitProgram(mp, program, datasetId, conf.Language, runId, false)
	return c.RenderJson(map[string]interface{}{
		"status":  "success",
		"runId":   runId,
		"attempt": "Attempt submitted",
	})
}

func (c SecuredApplication) RevertProgram(mpNumString string) revel.Result {

	var mp models.MachineProblem

	mpNum, err := strconv.Atoi(mpNumString)
	if err != nil {
		c.Flash.Error("Invalid machine problem")
		return c.Render(routes.PublicApplication.Index())
	}

	user := c.connected()
	stats.Incr("App", "ProgramRevert")

	if mp, err = models.FindOrCreateMachineProblemByUser(user, mpNum); err != nil {
		return c.RenderJson(map[string]interface{}{
			"status": "error",
			"data":   "Could not create mp.",
		})
	}

	if prog, err := revertMachineProblemProgram(user, mp); err == nil {
		return c.RenderJson(map[string]interface{}{
			"status":  "success",
			"program": prog.Text,
		})
	} else {
		return c.RenderJson(map[string]interface{}{
			"status": "error",
		})
	}
}

func (c SecuredApplication) SaveProgram(mpNumString string) revel.Result {
	var program string
	var mp models.MachineProblem

	mpNum, err := strconv.Atoi(mpNumString)
	if err != nil {
		c.Flash.Error("Invalid machine problem")
		return c.Render(routes.PublicApplication.Index())
	}

	user := c.connected()

	c.Params.Bind(&program, "program")

	stats.Incr("App", "ProgramSave")

	if mp, err = models.FindOrCreateMachineProblemByUser(user, mpNum); err != nil {
		return c.RenderJson(map[string]interface{}{
			"status": "error",
			"data":   "Could not create mp.",
		})
	}

	if prog, err := models.CreateProgram(mp, program); err != nil {
		return c.RenderJson(map[string]interface{}{
			"status": "error",
			"data":   "Could not save program.",
		})
	} else {
		return c.RenderJson(map[string]interface{}{
			"status": "success",
			"id":     prog.Id,
		})
	}
}

func (c SecuredApplication) Program(programIdString string) revel.Result {

	programId, err := strconv.Atoi(programIdString)
	if err != nil {
		c.Flash.Error("Invalid program id")
		return c.Render(routes.PublicApplication.Index())
	}

	user := c.connected()

	prog, err := models.FindProgram(int64(programId))
	if err != nil {
		return c.RenderText("error")
	}
	mp, err := models.FindMachineProblem(prog.MachineProblemInstanceId)
	if err != nil || mp.UserInstanceId != user.Id {
		return c.RenderText("error")
	} else {
		c.RenderArgs["program"] = prog
		return c.Render()
	}
}

func (c SecuredApplication) ProgramDiff(programIdString1 string, programIdString2 string) revel.Result {

	programId1, err := strconv.Atoi(programIdString1)
	if err != nil {
		c.Flash.Error("Invalid program id")
		return c.Render(routes.PublicApplication.Index())
	}
	programId2, err := strconv.Atoi(programIdString2)
	if err != nil {
		c.Flash.Error("Invalid program id")
		return c.Render(routes.PublicApplication.Index())
	}

	user := c.connected()

	prog1, err1 := models.FindProgram(int64(programId1))
	prog2, err2 := models.FindProgram(int64(programId2))

	if err1 != nil || err2 != nil {
		return c.RenderText("error")
	}
	mp1, err1 := models.FindMachineProblem(prog1.MachineProblemInstanceId)
	mp2, err2 := models.FindMachineProblem(prog2.MachineProblemInstanceId)
	if err1 != nil || err2 != nil || mp1.Id != mp2.Id || mp1.UserInstanceId != user.Id || mp2.UserInstanceId != user.Id {
		return c.RenderText("error")
	}

	c.RenderArgs["program1"] = prog1
	c.RenderArgs["program2"] = prog2
	return c.Render()
}

func (c SecuredApplication) RecentProgram(mpNumString string) revel.Result {
	var mp models.MachineProblem

	mpNum, err := strconv.Atoi(mpNumString)
	if err != nil {
		c.Flash.Error("Invalid machine problem")
		return c.Render(routes.PublicApplication.Index())
	}

	user := c.connected()

	if mp, err = models.FindOrCreateMachineProblemByUser(user, mpNum); err != nil {
		return c.RenderText("Cannot find MP")
	}

	if prog, err := machineProblemProgram(user, mp); err != nil {
		return c.RenderText("Cannot find program")
	} else {
		return c.RenderText(prog.Text)
	}
}

type programListSummary struct {
	Id      int64     `json:"id"`
	Snippet string    `json:"snippet"`
	Created time.Time `json:"created"`
}

func summarizeProgram(prog string) string {
	s := strings.Replace(strings.TrimSpace(prog), "\n", "  ", -1)
	bs := []byte(s)
	if len(bs) < 50 {
		return string(bs)
	} else {
		return string(bs[:50])
	}
}

func recentPrograms(user models.User, mp models.MachineProblem, count int) ([]programListSummary, error) {

	progs, err := models.RecentPrograms(mp, count)
	if err != nil {
		return nil, err
	}

	progListSummary := make([]programListSummary, len(progs))
	for i := range progs {
		progListSummary[i] = programListSummary{
			Id:      progs[i].Id,
			Snippet: summarizeProgram(progs[i].Text),
			Created: progs[i].Created,
		}
	}

	return progListSummary, nil
}

func (c SecuredApplication) RecentPrograms(mpNumString string, countString string) revel.Result {
	var mp models.MachineProblem

	mpNum, err := strconv.Atoi(mpNumString)
	if err != nil {
		c.Flash.Error("Invalid machine problem")
		return c.Render(routes.PublicApplication.Index())
	}

	count, _ := strconv.Atoi(countString)

	user := c.connected()

	if mp, err = models.FindOrCreateMachineProblemByUser(user, mpNum); err != nil {
		return c.RenderJson(map[string]interface{}{
			"status": "error",
			"data":   "Failed to find machine problem",
		})
	}

	if progs, err := recentPrograms(user, mp, count); err != nil {
		return c.RenderJson(map[string]interface{}{
			"status": "error",
			"data":   "Failed to find recent programs",
		})
	} else {
		return c.RenderJson(map[string]interface{}{
			"status": "success",
			"data":   progs,
		})
	}

}

func (c SecuredApplication) Attempts(mpNumString string) revel.Result {
	var mp models.MachineProblem

	mpNum, err := strconv.Atoi(mpNumString)
	if err != nil {
		c.Flash.Error("Invalid machine problem")
		return c.Render(routes.PublicApplication.Index())
	}

	user := c.connected()

	if mp, err = models.FindOrCreateMachineProblemByUser(user, mpNum); err != nil {
		c.Flash.Error("Cannot query machine problem number")
		return c.Render(routes.PublicApplication.Index())
	}

	c.RenderArgs["mp_attempts"] = getAttemptsSummary(user, mp)

	conf, _ := ReadMachineProblemConfig(mpNum)
	c.RenderArgs["title"] = conf.Name + " Attempts"

	return c.Render()
}

func (c SecuredApplication) Attempt(attemptIdString string) revel.Result {

	attemptId, err := strconv.Atoi(attemptIdString)
	if err != nil {
		c.Flash.Error("Invalid attempt Id")
		return c.Render(routes.PublicApplication.Index())
	}

	user := c.connected()

	attempt, err := models.FindAttempt(int64(attemptId))
	if err != nil {
		c.Flash.Error("Cannot query attempt id")
		return c.Render(routes.PublicApplication.Index())
	}

	mp, err := models.FindMachineProblem(attempt.MachineProblemInstanceId)
	if err != nil || mp.UserInstanceId != user.Id {
		c.Flash.Error("Cannot query attempt id")
		return c.Render(routes.PublicApplication.Index())
	}

	attempt_c_data := new(server.InternalCData)

	if err := json.Unmarshal([]byte(attempt.InternalCData), attempt_c_data); err == nil {
		c.RenderArgs["attempt_c_data"] = attempt_c_data
	}

	conf, _ := ReadMachineProblemConfig(mp.Number)
	if !MachineProblemCodingDeadlineExpiredQ(conf) {
		c.RenderArgs["show_grade_button"] = true
	}

	prog, err := models.FindProgram(attempt.ProgramInstanceId)
	if err != nil {
		c.Flash.Error("Cannot find program.")
		return c.Render(routes.PublicApplication.Index())
	}

	c.RenderArgs["mp"] = mp
	c.RenderArgs["title"] = conf.Name + " Attempt"

	c.RenderArgs["mp_config"] = conf
	c.RenderArgs["attempt"] = attempt
	c.RenderArgs["program"] = prog

	qs, err := models.FindQuestionsByMachineProblem(mp)
	if err != nil {
		c.Flash.Error("Invalid questions")
		return c.Render(routes.PublicApplication.Index())
	}

	q, err := models.FindQuestionItems(mp, qs)
	if err != nil {
		c.Flash.Error("Invalid question items")
		return c.Render(routes.PublicApplication.Index())
	} else {
		c.RenderArgs["questions"] = q
	}

	if attemptFailed(attempt) {
		c.RenderArgs["status"] = attemptFailedMessage(attempt)
	} else {
		c.RenderArgs["status"] = "Correct solution for this dataset."
	}

	return c.Render()
}

func (c SecuredApplication) AttemptRunInformation(mpNumString string, runIdString string) revel.Result {

	runId := runIdString

	attempt, err := models.FindAttemptByRunId(runId)

	if err != nil {
		return c.RenderJson(map[string]interface{}{
			"status": "error-not-found",
			"title":  "Did not finish running attempt.",
			"data":   "Cannot find attempt with run id = " + runIdString,
		})
	} else if attemptFailed(attempt) {
		return c.RenderJson(map[string]interface{}{
			"status": "error",
			"title":  attemptFailedReason(attempt),
			"data":   attemptFailedMessage(attempt),
			"link":   fmt.Sprint("/attempt/", attempt.Id),
		})
	} else {
		return c.RenderJson(map[string]interface{}{
			"status": "success",
			"data":   attempt,
			"id":     attempt.Id,
		})
	}
}

func (c SecuredApplication) SaveQuestion(mpNumString string, questionNumString string) revel.Result {
	var answer string

	mpNum, err := strconv.Atoi(mpNumString)
	questionNum, err := strconv.Atoi(questionNumString)

	c.Params.Bind(&answer, "answer")

	if answer == "" {
		return c.RenderJson(map[string]interface{}{
			"status": "error",
			"data":   "Blank answer was not saved.",
		})
	}

	user := c.connected()

	mp, err := models.FindOrCreateMachineProblemByUser(user, mpNum)
	if err != nil {
		return c.RenderJson(map[string]interface{}{
			"status": "error",
			"data":   "Cannot find machine problem.",
		})
	}

	conf, _ := ReadMachineProblemConfig(mp.Number)
	if MachineProblemCodingDeadlineExpiredQ(conf) {
		return c.RenderJson(map[string]interface{}{
			"status": "error-deadline",
			"data":   "Coding deadline has passed.",
		})
	}

	questions, err := models.FindOrCreateQuestionsByMachineProblem(mp)
	if err != nil {
		return c.RenderJson(map[string]interface{}{
			"status": "error",
			"data":   "Error: create questions.",
		})
	}

	err = models.SaveQuestion(mp, questions, int64(questionNum), answer)

	stats.Incr("App", "QuestionSave")

	return c.RenderJson(map[string]interface{}{
		"status": "success",
		"data":   answer,
	})
}

func (c SecuredApplication) ShowGrade(mpNumString string) revel.Result {

	mpNum, err := strconv.Atoi(mpNumString)
	if err != nil {
		c.Flash.Error("Invalid mp Id")
		return c.Render(routes.PublicApplication.Index())
	}

	user := c.connected()

	mp, err := models.FindOrCreateMachineProblemByUser(user, mpNum)
	if err != nil {
		c.Flash.Error("Invalid mp Id")
		return c.Render(routes.PublicApplication.Index())
	}

	grade, err := models.FindGradeByMachineProblem(mp)
	if err != nil {
		c.Flash.Error("Invalid grade Id")
		return c.Render(routes.PublicApplication.Index())
	}

	return c.Grade(strconv.Itoa(int(grade.Id)))
}

func (c SecuredApplication) Grade(gradeIdString string) revel.Result {

	gradeId, err := strconv.Atoi(gradeIdString)
	if err != nil {
		c.Flash.Error("Invalid grade Id")
		return c.Render(routes.PublicApplication.Index())
	}

	grade, err := models.FindGrade(int64(gradeId))
	if err != nil {
		c.Flash.Error("Invalid grade Id")
		return c.Render(routes.PublicApplication.Index())
	}

	attempt, err := models.FindAttempt(grade.AttemptInstanceId)
	if err != nil || attempt.Id == 0 {
		c.Flash.Error("Invalid attempt")
		return c.Render(routes.PublicApplication.Index())
	}

	mp, err := models.FindMachineProblem(attempt.MachineProblemInstanceId)
	if err != nil {
		c.Flash.Error("Invalid mp")
		return c.Render(routes.PublicApplication.Index())
	}

	c.RenderArgs["grade"] = grade
	if grade.Reasons != "" {
		c.RenderArgs["reasons"] = strings.Split(grade.Reasons, ",")
	}

	c.RenderArgs["mp"] = mp

	conf, _ := ReadMachineProblemConfig(mp.Number)
	c.RenderArgs["title"] = conf.Name + " Grade"

	c.RenderArgs["grade"] = grade

	if qs, err := models.FindQuestionsByMachineProblem(mp); err == nil {
		if q, err := models.FindQuestionItems(mp, qs); err == nil {
			c.RenderArgs["questions"] = q
		}
	}

	c.RenderArgs["attempt"] = attempt

	if prog, err := models.FindProgram(attempt.ProgramInstanceId); err == nil {
		c.RenderArgs["program"] = prog
	}

	if grade.PeerReviewScore > 0 {
		if pr, err := models.FindPeerReviewsWithGrade(grade); err == nil {
			c.RenderArgs["peer_reviews"] = pr
		}
	}

	return c.Render()
}

func (c SecuredApplication) ComputeGrade(attemptIdString string) revel.Result {

	attemptId, err := strconv.Atoi(attemptIdString)
	if err != nil {
		c.Flash.Error("Invalid attempt Id")
		return c.Render(routes.PublicApplication.Index())
	}

	attempt, err := models.FindAttempt(int64(attemptId))
	if err != nil || attempt.Id == 0 {
		return c.RenderJson(map[string]interface{}{
			"status": "error",
			"title":  "Failed to grade machine problem",
			"data":   "Cannot find attempt " + attemptIdString + ".",
		})
	}

	mp, err := models.FindMachineProblem(attempt.MachineProblemInstanceId)
	if err != nil || mp.Id == 0 {
		return c.RenderJson(map[string]interface{}{
			"status": "error",
			"title":  "Failed to grade",
			"data":   "Cannot find machine problem instance.",
		})
	}

	user := c.connected()

	if mp.UserInstanceId != user.Id {
		return c.RenderJson(map[string]interface{}{
			"status": "error",
			"title":  "Failed to grade",
			"data":   "Invalid user.",
		})
	}

	runId := generateRunId(user)

	conf, _ := ReadMachineProblemConfig(mp.Number)

	stats.Incr("App", "GradeSubmission")

	prog, _ := models.FindProgram(attempt.ProgramInstanceId)
	server.SubmitProgram(mp, prog.Text, -1, conf.Language, runId, true)

	return c.RenderJson(map[string]interface{}{
		"status":  "success",
		"mpId":    strconv.Itoa(int(mp.Id)),
		"runId":   runId,
		"attempt": "Grade submitted",
	})
}

func (c SecuredApplication) GradeRunInformation(attemptIdString string, runIdString string) revel.Result {

	attemptId, err := strconv.Atoi(attemptIdString)
	if err != nil {
		c.Flash.Error("Invalid attempt Id")
		return c.Render(routes.PublicApplication.Index())
	}

	user := c.connected()

	attempt, err := models.FindAttempt(int64(attemptId))
	if err != nil {
		return c.RenderJson(map[string]interface{}{
			"status": "error-not-found",
			"title":  "Did not finish running attempt.",
			"data": "Cannot find attempt with attempt id = " +
				attemptIdString + " ... " +
				fmt.Sprint(err),
		})
	}

	mp, err := models.FindMachineProblem(attempt.MachineProblemInstanceId)
	if err != nil {
		return c.RenderJson(map[string]interface{}{
			"status": "error-not-found",
			"title":  "Did not finish running attempt.",
			"data": "Cannot find mp with mp id = " +
				fmt.Sprint(attempt.MachineProblemInstanceId, " .... ", err),
		})
	}

	grade, err := models.FindGradeByRunId(runIdString)
	if err != nil {
		//revel.TRACE.Println("did not find grade...")
		return c.RenderJson(map[string]interface{}{
			"status": "error-not-found",
			"title":  "Did not finish running attempt.",
			"data": "Cannot find mp with grade with runid  = " +
				runIdString + " ... " + fmt.Sprint(err),
		})
	}

	if !models.AllGraded(grade) {
		//revel.TRACE.Println("not all graded...")
		return c.RenderJson(map[string]interface{}{
			"status": "error-not-found",
			"title":  "Did not finish grading.",
			"data":   "Cannot display incomplete grade for mp = " + strconv.Itoa(int(mp.Number)),
		})
	} else if err := doPostCourseraGrade(user, mp, grade, "code", false); err != nil {
		return c.RenderJson(map[string]interface{}{
			"status": "success",
			"title":  "Info: Did not save grade to Coursera.",
			"data":   "Make sure you are connected to Coursera",
			"id":     grade.Id,
		})
	}

	grade, err = models.FindGradeByRunId(runIdString)

	return c.RenderJson(map[string]interface{}{
		"status": "success",
		"data":   grade,
		"id":     grade.Id,
	})
}

func (c SecuredApplication) PeerReview(mpNumString string) revel.Result {

	mpNum, err := strconv.Atoi(mpNumString)
	if err != nil {
		c.Flash.Error("Invalid mp Id")
		return c.Render(routes.PublicApplication.Index())
	}

	user := c.connected()

	mp, err := models.FindOrCreateMachineProblemByUser(user, mpNum)
	if err != nil {
		c.Flash.Error("Invalid mp Id")
		return c.Render(routes.PublicApplication.Index())
	}

	conf, _ := ReadMachineProblemConfig(mpNum)

	c.RenderArgs["mp_config"] = conf
	c.RenderArgs["title"] = conf.Name + " Peer Review"

	if MachineProblemPeerReviewDeadlineExpiredQ(conf) {
		c.RenderArgs["past_deadline"] = true
	} else if MachineProblemCodingDeadlineExpiredQ(conf) {

		reviews, err := models.GetPeerReviewsByReviewerAndMachineProblem(user, mp)

		if err != nil || len(reviews) < 3 {
			ii := 0
			for len(reviews) < 3 && ii < 100 {
				if rev, err := models.CreatePeerReviewWithReviewerAndMachineProblem(user, mp); err == nil {
					reviews = append(reviews, rev)
				}
				ii++
			}
		}

		type reviewT struct {
			Index         int
			Review        models.PeerReview
			Questions     models.Questions
			QuestionItems []models.QuestionItem
			Program       models.Program
		}

		var res []reviewT

		for _, r := range reviews {
			if grade, err := models.FindGrade(r.GradeInstanceId); err == nil {
				if attempt, err := models.FindAttempt(grade.AttemptInstanceId); err == nil {
					program, _ := models.FindProgram(attempt.ProgramInstanceId)
					tmp, _ := models.FindMachineProblem(grade.MachineProblemId)
					q, _ := models.FindQuestionsByMachineProblem(tmp)
					qs, _ := models.FindQuestionItems(mp, q)
					k := reviewT{
						Index:         len(res) + 1,
						Review:        r,
						Questions:     q,
						QuestionItems: qs,
						Program:       program,
					}
					res = append(res, k)
				}
			}
		}

		c.RenderArgs["reviews"] = res
	} else {
		c.RenderArgs["still_coding"] = true
	}

	return c.Render()
}

type PeerReviewRecord struct {
	Id              int64  `json:"id"`
	QuestionScore   int64  `json:"question_score"`
	QuestionComment string `json:"question_comment"`
	CodeScore       int64  `json:"code_score"`
	CodeComment     string `json:"code_comment"`
}

func (c SecuredApplication) UpdatePeerReview(mpNumString string) revel.Result {
	mpNum, err := strconv.Atoi(mpNumString)
	if err != nil {
		return c.RenderJson(map[string]interface{}{
			"status": "error",
			"title":  "Cannot Find Machine Problem",
			"data":   "Cannot find machine problem --- invalid MP Id.",
			"error":  err,
		})
	}

	user := c.connected()

	mp, err := models.FindOrCreateMachineProblemByUser(user, mpNum)
	if err != nil {
		return c.RenderJson(map[string]interface{}{
			"status": "error",
			"title":  "Cannot Find Machine Problem",
			"data":   "Cannot find machine problem.",
			"error":  err,
		})
	}

	if _, err := models.FindLastAttemptByMachineProblem(mp); err != nil {
		if _, err = models.CreateDummyAttemptByMachineProblem(mp); err != nil {
			return c.RenderJson(map[string]interface{}{
				"status": "error",
				"title":  "Error: Cannot Submit Peer Review.",
				"data":   "We require that you have attempted the MP to submit a peer review.",
				"error":  err,
			})
		}
	}

	var reviewsString string
	c.Params.Bind(&reviewsString, "reviews")

	//revel.TRACE.Println(reviewsString)

	var reviews []PeerReviewRecord
	if err := json.Unmarshal([]byte(reviewsString), &reviews); err != nil {
		return c.RenderJson(map[string]interface{}{
			"status": "error",
			"title":  "Error: Cannot Parse Review Request.",
			"data":   "The system was not able to parse your review request.",
			"error":  err,
		})
	}

	prs := make([]models.PeerReview, len(reviews))
	for ii, review := range reviews {
		pr, err := models.FindPeerReview(review.Id)
		if err != nil || pr.Reviewer != user.Id {
			return c.RenderJson(map[string]interface{}{
				"status": "error",
				"title":  "Error: Cannot Submit Review.",
				"data":   "Cannot submit review request -- invalid reviewer.",
			})
		}
		if strings.TrimSpace(review.QuestionComment) == "" ||
			strings.TrimSpace(review.CodeComment) == "" {
			return c.RenderJson(map[string]interface{}{
				"status": "error",
				"title":  "Error: Cannot Submit Empty Review.",
				"data":   "Cannot submit review. Please make sure to offer useful comments to your peers.",
			})
		}
		prs[ii] = pr
	}

	for ii, pr := range prs {
		review := reviews[ii]
		_, err := models.UpdatePeerReview(pr, review.CodeScore, review.CodeComment,
			review.QuestionScore, review.QuestionComment)
		if err != nil {
			return c.RenderJson(map[string]interface{}{
				"status": "error",
				"title":  "Error: Cannot Save Review",
				"data":   "The system was not able to submit your review to the database. Report this issue if problem presists.",
			})
		}
	}

	var grade models.Grade
	if grade, err = models.UpdateGradePeerReview(user, mp); err != nil {
		return c.RenderJson(map[string]interface{}{
			"status": "error",
			"title":  "Error: Peer Review Grade was not Updated.",
			"data": "The system was not able to update your peer review grade. " +
				"Report this issue if problem presists. " +
				fmt.Sprint(err),
		})
	}

	if err := doPostCourseraGrade(user, mp, grade, "peer", false); err != nil {
		return c.RenderJson(map[string]interface{}{
			"status": "error",
			"title":  "Error: Did not save grade to Coursera.",
			"data":   "Make sure you are connected to Coursera",
		})
	}

	link := "/grade/" + strconv.Itoa(int(grade.Id))
	return c.RenderJson(map[string]interface{}{
		"status": "success",
		"title":  "Peer Review has been Submitted",
		"data": "Review has been logged and submitted. Make sure to visit the <a href=\"" +
			link +
			"\"> machine problem grade page</a> and re-submit your grade to Coursera.",
		"link": link,
	})
}

func (c SecuredApplication) WorkerMessages() revel.Result {
	c.RenderArgs["status"] = stats.Packets
	return c.Render()
}

func (c SecuredApplication) WorkerStatus() revel.Result {
	if WorkerIsAlive() {
		c.RenderArgs["alive"] = true
		return c.Render()
	} else {
		return c.WorkerMessages()
	}
}

func (c SecuredApplication) ResetCache() revel.Result {
	ResetApplication()
	return c.RenderJson(map[string]interface{}{
		"status": "success",
	})
}

func (c SecuredApplication) DeleteTemporaryFiles() revel.Result {
	return c.RenderJson(map[string]interface{}{
		"status": "success",
	})
}

func (c SecuredApplication) QuestionHistory(mpNumString string, questionNumString string) revel.Result {
	mpNum, err := strconv.Atoi(mpNumString)
	if err != nil {
		c.Flash.Error("Invalid mp number")
		return c.Render(routes.PublicApplication.Index())
	}
	questionNum, err := strconv.Atoi(questionNumString)
	if err != nil {
		c.Flash.Error("Invalid question number")
		return c.Render(routes.PublicApplication.Index())
	}

	user := c.connected()

	mp, err := models.FindOrCreateMachineProblemByUser(user, mpNum)
	if err != nil {
		c.Flash.Error("Invalid mp Id")
		return c.Render(routes.PublicApplication.Index())
	}

	qs, err := models.FindOrCreateQuestionsByMachineProblem(mp)
	if err != nil {
		c.Flash.Error("Invalid questions")
		return c.Render(routes.PublicApplication.Index())
	}

	q, err := models.FindQuestionHistory(qs, int64(questionNum))
	if err != nil {
		c.Flash.Error("Invalid question items")
		return c.Render(routes.PublicApplication.Index())
	}

	c.RenderArgs["mp"] = mp
	c.RenderArgs["number"] = questionNum
	c.RenderArgs["questions"] = q

	return c.Render()
}

func (c SecuredApplication) GradeHistory(mpNumString string) revel.Result {

	mpNum, err := strconv.Atoi(mpNumString)
	if err != nil {
		c.Flash.Error("Invalid mp number")
		return c.Render(routes.PublicApplication.Index())
	}

	user := c.connected()

	mp, err := models.FindOrCreateMachineProblemByUser(user, mpNum)
	if err != nil {
		c.Flash.Error("Invalid mp Id")
		return c.Render(routes.PublicApplication.Index())
	}

	grades, err := models.FindGradesByMachineProblem(mp)
	if err != nil {
		c.Flash.Error("Invalid grades items")
		return c.Render(routes.PublicApplication.Index())
	}
	c.RenderArgs["grades"] = grades

	return c.Render()
}

func ResetApplication() {
	if IsMaster {
		ResetWorkerCache()
	}
	ResetConfig()
	models.ResetModels()
	ResetControllers()
	stats.ResetStats()
	server.ResetServer()
}
