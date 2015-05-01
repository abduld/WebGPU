package controllers

import (
	"bytes"
	"io/ioutil"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"text/template"
	"time"

	"encoding/json"
	"errors"
	"wb/app/bigcode"
	. "wb/app/config"
	"wb/app/models"
	"wb/app/routes"
	"wb/app/server"

	"github.com/revel/revel"
)

type AdminApplication struct {
	SecuredApplication
}

const workerIsDeadEmailTemplate = `
Hello admin,

The worker node {{ .WorkerAddress }} is not responsive for the past {{ .Minutes }} minutes.

Note: This email might be a false positive and you can check the status of the worker by visiting

   http://www.webgpu.com/workers

If not listed, then you'd have to reboot the machine and start the worker daemon back up again.

Thank you.`

const taEmailTemplate = `
Hello {{.FirstName}} {{.LastName}} ({{.UserName}}),

The TA for the class has updated your grade. Please visit the grades page to see the update.
If you have any questions, please send a message to the TA.

Thank you.`

const hoursBetweenEmails = 1

var (
	WorkersResetEmailLog   map[string]time.Time = map[string]time.Time{}
	WorkersResetEmailCount map[string]int64     = map[string]int64{}
)

func SendAdminsEmailWhenWorkerIsDead(workerAddress string, minutes int) {
	type WorkerIsDeadEmailTemplateParams struct {
		WorkerAddress string
		Minutes       int
	}

	if SendAdminEmails == false {
		return
	}

	if _, ok := WorkersResetEmailCount[workerAddress]; !ok {
		WorkersResetEmailCount[workerAddress] = 1
	} else {
		WorkersResetEmailCount[workerAddress]++
	}

	if t, ok := WorkersResetEmailLog[workerAddress]; !ok || time.Since(t).Hours() > hoursBetweenEmails {
		WorkersResetEmailLog[workerAddress] = time.Now()

		if val, ok := WorkersResetEmailCount[workerAddress]; ok && val > 2 {
			return
		}

		templt := template.Must(template.New("workerIsDeadEmailTemplate").Parse(workerIsDeadEmailTemplate))

		params := WorkerIsDeadEmailTemplateParams{
			WorkerAddress: workerAddress,
			Minutes:       minutes,
		}

		var buf bytes.Buffer

		if err := templt.Execute(&buf, params); err != nil {
			revel.TRACE.Println(err)
			return
		}
		for _, email := range SiteAdmins {
			SendEmail(ApplicationEmail, email, "Unresponsive worker", string(buf.Bytes()))
		}
	}

}

type studentRecord struct {
	Id             int64                 `json:"id"`
	Student        models.User           `json:"student"`
	MachineProblem models.MachineProblem `json:"machine_problem"`
	Questions      []models.QuestionItem `json:"questions"`
	Grade          models.Grade          `json:"grade"`
	Attempt        models.Attempt        `json:"attempt"`
	Program        models.Program        `json:"program"`
}

func (c AdminApplication) MachineProblemClassRoster(mpNumString, startIdxString, endIdxString string) revel.Result {

	mpNum, err := strconv.Atoi(mpNumString)
	if err != nil {
		c.Flash.Error("Invalid mp Id")
		return c.Render(routes.PublicApplication.Index())
	}
	startIdx, err := strconv.Atoi(startIdxString)
	if err != nil {
		c.Flash.Error("Invalid start index")
		return c.Render(routes.PublicApplication.Index())
	}
	endIdx, err := strconv.Atoi(endIdxString)
	if err != nil {
		c.Flash.Error("Invalid end index")
		return c.Render(routes.PublicApplication.Index())
	}

	admin := c.connected()
	if !models.IsAdmin(admin) {
		c.Flash.Error("No admin privilages")
		return c.Render(routes.PublicApplication.Index())
	}

	conf, _ := ReadMachineProblemConfig(mpNum)

	c.RenderArgs["admin"] = admin
	c.RenderArgs["mp_num"] = mpNum
	c.RenderArgs["mp_config"] = conf

	students, err := models.FindUsersByMachineProblemNumber(mpNum)
	if err != nil {
		c.Flash.Error("Invalid mp number")
		return c.Render(routes.PublicApplication.Index())
	}
	if startIdx < 0 {
		startIdx = 0
	}
	if endIdx > len(students) || endIdx == 0 {
		endIdx = len(students)
	}
	students = students[startIdx:endIdx]

	class_roster := make([]studentRecord, len(students))
	for idx, student := range students {
		mp, _ := models.FindMachineProblemByUser(student, mpNum)
		grade, _ := models.FindGradeByMachineProblem(mp)

		//qs, _ := models.FindQuestionsByMachineProblem(mp)
		//qis, _ := models.FindQuestionItems(mp, qs)
		class_roster[idx] = studentRecord{
			Id:             student.Id,
			Student:        student,
			MachineProblem: mp,
			//Questions:      qis,
			Grade: grade,
		}
	}

	c.RenderArgs["class_roster"] = class_roster

	return c.Render()
}

func (c AdminApplication) ShowStudentMachineProblem(studentIdString string, mpNumString string) revel.Result {

	studentId, err := strconv.Atoi(studentIdString)
	if err != nil {
		c.Flash.Error("Invalid user Id")
		return c.Render(routes.PublicApplication.Index())
	}
	mpNum, err := strconv.Atoi(mpNumString)
	if err != nil {
		c.Flash.Error("Invalid mp Id")
		return c.Render(routes.PublicApplication.Index())
	}

	admin := c.connected()
	if !models.IsAdmin(admin) {
		c.Flash.Error("No admin privilages")
		return c.Render(routes.PublicApplication.Index())
	}

	conf, _ := ReadMachineProblemConfig(mpNum)

	student, err := models.FindUser(int64(studentId))
	if err != nil {
		c.Flash.Error("Cannot find user")
		return c.Render(routes.PublicApplication.Index())
	}

	mp, _ := models.FindMachineProblemByUser(student, mpNum)
	grade, _ := models.FindGradeByMachineProblem(mp)
	attempt, _ := models.FindAttempt(grade.AttemptInstanceId)
	program, _ := models.FindProgram(attempt.ProgramInstanceId)
	qs, _ := models.FindQuestionsByMachineProblem(mp)
	qis, _ := models.FindQuestionItems(mp, qs)

	c.RenderArgs["admin"] = admin
	c.RenderArgs["mp_num"] = mpNum
	c.RenderArgs["mp_config"] = conf
	c.RenderArgs["student"] = student
	c.RenderArgs["mp"] = mp
	c.RenderArgs["attempt"] = attempt
	c.RenderArgs["program"] = program
	c.RenderArgs["questions"] = qis
	c.RenderArgs["grade"] = grade
	bg, err := getBigCodeSuggestion(mpNum, program)
	if err != nil {
		revel.INFO.Println("Failed to get big code suggestion  :::  ", err)
	} else if len(bg) != 3 {
		revel.INFO.Println("The number of suggestions recieved was  :::  ", len(bg), " was expecting 3")
	} else {
		//revel.INFO.Println("Got bigcode suggestions")
		//revel.INFO.Println("min  = ", bg[0])

		c.RenderArgs["bigcode"] = bg
		c.RenderArgs["bigcode_min"] = bg[0]
		c.RenderArgs["bigcode_max"] = bg[1]
		c.RenderArgs["bigcode_random"] = bg[2]

		randomSelection := rand.Intn(3)
		switch randomSelection {
		case 0:
			c.RenderArgs["bigcode_min_active"] = "active"
			c.RenderArgs["bigcode_max_active"] = ""
			c.RenderArgs["bigcode_random_active"] = ""
		case 1:
			c.RenderArgs["bigcode_min_active"] = ""
			c.RenderArgs["bigcode_max_active"] = "active"
			c.RenderArgs["bigcode_random_active"] = ""
		case 2:
			c.RenderArgs["bigcode_min_active"] = ""
			c.RenderArgs["bigcode_max_active"] = ""
			c.RenderArgs["bigcode_random_active"] = "active"
		}
	}

	return c.Render()
}

func (c AdminApplication) ShowStudentPrograms(studentIdString string, mpNumString string) revel.Result {

	studentId, err := strconv.Atoi(studentIdString)
	if err != nil {
		c.Flash.Error("Invalid user Id")
		return c.Render(routes.PublicApplication.Index())
	}
	mpNum, err := strconv.Atoi(mpNumString)
	if err != nil {
		c.Flash.Error("Invalid mp Id")
		return c.Render(routes.PublicApplication.Index())
	}

	admin := c.connected()
	if !models.IsAdmin(admin) {
		c.Flash.Error("No admin privilages")
		return c.Render(routes.PublicApplication.Index())
	}

	student, err := models.FindUser(int64(studentId))
	if err != nil {
		c.Flash.Error("Cannot find user")
		return c.Render(routes.PublicApplication.Index())
	}

	mp, _ := models.FindMachineProblemByUser(student, mpNum)
	programs, _ := models.FindProgramsByMachineProblem(mp)

	c.RenderArgs["mp_num"] = mpNum
	c.RenderArgs["mp"] = mp
	c.RenderArgs["student"] = student
	c.RenderArgs["programs"] = programs

	return c.Render()
}

func (c AdminApplication) ShowStudentProgram(studentIdString string, mpNumString string, programIdString string) revel.Result {

	studentId, err := strconv.Atoi(studentIdString)
	if err != nil {
		c.Flash.Error("Invalid user Id")
		return c.Render(routes.PublicApplication.Index())
	}
	mpNum, err := strconv.Atoi(mpNumString)
	if err != nil {
		c.Flash.Error("Invalid mp Id")
		return c.Render(routes.PublicApplication.Index())
	}

	programId, err := strconv.Atoi(programIdString)
	if err != nil {
		c.Flash.Error("Invalid program Id")
		return c.Render(routes.PublicApplication.Index())
	}

	admin := c.connected()
	if !models.IsAdmin(admin) {
		c.Flash.Error("No admin privilages")
		return c.Render(routes.PublicApplication.Index())
	}

	student, err := models.FindUser(int64(studentId))
	if err != nil {
		c.Flash.Error("Cannot find user")
		return c.Render(routes.PublicApplication.Index())
	}

	mp, _ := models.FindMachineProblemByUser(student, mpNum)
	program, _ := models.FindProgram(int64(programId))

	c.RenderArgs["mp_num"] = mpNum
	c.RenderArgs["mp"] = mp
	c.RenderArgs["student"] = student
	c.RenderArgs["program"] = program

	return c.Render()
}

func (c AdminApplication) ShowStudentAttempts(studentIdString string, mpNumString string) revel.Result {

	studentId, err := strconv.Atoi(studentIdString)
	if err != nil {
		c.Flash.Error("Invalid user Id")
		return c.Render(routes.PublicApplication.Index())
	}
	mpNum, err := strconv.Atoi(mpNumString)
	if err != nil {
		c.Flash.Error("Invalid mp Id")
		return c.Render(routes.PublicApplication.Index())
	}

	admin := c.connected()
	if !models.IsAdmin(admin) {
		c.Flash.Error("No admin privilages")
		return c.Render(routes.PublicApplication.Index())
	}

	student, err := models.FindUser(int64(studentId))
	if err != nil {
		c.Flash.Error("Cannot find user")
		return c.Render(routes.PublicApplication.Index())
	}

	mp, _ := models.FindMachineProblemByUser(student, mpNum)
	attempts, _ := models.FindAttemptsByMachineProblem(mp)

	c.RenderArgs["mp_num"] = mpNum
	c.RenderArgs["mp"] = mp
	c.RenderArgs["student"] = student
	c.RenderArgs["attempts"] = attempts

	return c.Render()
}

func (c AdminApplication) ShowStudentAttempt(studentIdString string, mpNumString string, attemptIdString string) revel.Result {

	studentId, err := strconv.Atoi(studentIdString)
	if err != nil {
		c.Flash.Error("Invalid user Id")
		return c.Render(routes.PublicApplication.Index())
	}
	mpNum, err := strconv.Atoi(mpNumString)
	if err != nil {
		c.Flash.Error("Invalid mp Id")
		return c.Render(routes.PublicApplication.Index())
	}
	attemptId, err := strconv.Atoi(attemptIdString)
	if err != nil {
		c.Flash.Error("Invalid attempt Id")
		return c.Render(routes.PublicApplication.Index())
	}

	admin := c.connected()
	if !models.IsAdmin(admin) {
		c.Flash.Error("No admin privilages")
		return c.Render(routes.PublicApplication.Index())
	}

	student, err := models.FindUser(int64(studentId))
	if err != nil {
		c.Flash.Error("Cannot find user")
		return c.Render(routes.PublicApplication.Index())
	}

	mp, _ := models.FindMachineProblemByUser(student, mpNum)
	attempt, _ := models.FindAttempt(int64(attemptId))
	program, _ := models.FindProgram(attempt.ProgramInstanceId)
	qs, _ := models.FindQuestionsByMachineProblem(mp)
	qis, _ := models.FindQuestionItems(mp, qs)

	c.RenderArgs["mp_num"] = mpNum
	c.RenderArgs["mp"] = mp
	c.RenderArgs["student"] = student
	c.RenderArgs["attempt"] = attempt
	c.RenderArgs["program"] = program
	c.RenderArgs["questions"] = qis

	return c.Render()
}

func (c AdminApplication) ShowStudentGrades(studentIdString string, mpNumString string) revel.Result {

	studentId, err := strconv.Atoi(studentIdString)
	if err != nil {
		c.Flash.Error("Invalid user Id")
		return c.Render(routes.PublicApplication.Index())
	}
	mpNum, err := strconv.Atoi(mpNumString)
	if err != nil {
		c.Flash.Error("Invalid mp Id")
		return c.Render(routes.PublicApplication.Index())
	}

	admin := c.connected()
	if !models.IsAdmin(admin) {
		c.Flash.Error("No admin privilages")
		return c.Render(routes.PublicApplication.Index())
	}

	student, err := models.FindUser(int64(studentId))
	if err != nil {
		c.Flash.Error("Cannot find user")
		return c.Render(routes.PublicApplication.Index())
	}

	mp, _ := models.FindMachineProblemByUser(student, mpNum)
	grades, _ := models.FindGradesByMachineProblem(mp)

	c.RenderArgs["mp_num"] = mpNum
	c.RenderArgs["mp"] = mp
	c.RenderArgs["student"] = student
	c.RenderArgs["grades"] = grades

	return c.Render()
}

func (c AdminApplication) ShowStudentGrade(studentIdString string, mpNumString string, gradeIdString string) revel.Result {

	studentId, err := strconv.Atoi(studentIdString)
	if err != nil {
		c.Flash.Error("Invalid user Id")
		return c.Render(routes.PublicApplication.Index())
	}
	mpNum, err := strconv.Atoi(mpNumString)
	if err != nil {
		c.Flash.Error("Invalid mp Id")
		return c.Render(routes.PublicApplication.Index())
	}
	gradeId, err := strconv.Atoi(gradeIdString)
	if err != nil {
		c.Flash.Error("Invalid grade Id")
		return c.Render(routes.PublicApplication.Index())
	}

	admin := c.connected()
	if !models.IsAdmin(admin) {
		c.Flash.Error("No admin privilages")
		return c.Render(routes.PublicApplication.Index())
	}

	student, err := models.FindUser(int64(studentId))
	if err != nil {
		c.Flash.Error("Cannot find user")
		return c.Render(routes.PublicApplication.Index())
	}

	mp, _ := models.FindMachineProblemByUser(student, mpNum)
	grade, _ := models.FindGrade(int64(gradeId))
	attempt, _ := models.FindAttempt(grade.AttemptInstanceId)
	program, _ := models.FindProgram(attempt.ProgramInstanceId)
	qs, _ := models.FindQuestionsByMachineProblem(mp)
	qis, _ := models.FindQuestionItems(mp, qs)

	c.RenderArgs["mp_num"] = mpNum
	c.RenderArgs["mp"] = mp
	c.RenderArgs["student"] = student
	c.RenderArgs["grade"] = grade
	c.RenderArgs["attempt"] = attempt
	c.RenderArgs["program"] = program
	c.RenderArgs["questions"] = qis
	bg, err := getBigCodeSuggestion(mpNum, program)
	if err != nil {
		revel.INFO.Println("Failed to get big code suggestion  :::  ", err)
	} else {
		revel.INFO.Println("Got bigcode suggestions")
		c.RenderArgs["bigcode"] = bg
		c.RenderArgs["bigcode_min"] = bg[0]
		c.RenderArgs["bigcode_max"] = bg[1]
		c.RenderArgs["bigcode_random"] = bg[2]
	}

	return c.Render()
}

func getBigCodeSuggestion(mpNum int, program models.Program) ([]bigcode.BigCodeDatabaseEntry, error) {
	res := []bigcode.BigCodeDatabaseEntry{}
	mpNumString := strconv.Itoa(mpNum)

	if len(server.Workers) == 0 {
		return res, errors.New("No workers were found to get the code suggestion")
	}
	var workers []*server.WorkerInfo
	for _, k := range server.Workers {
		workers = append(workers, k)
	}
	worker := workers[rand.Intn(len(server.Workers))]
	js, err := json.Marshal(program)
	if err != nil {
		revel.INFO.Println(err)
		return res, err
	}
	b := bytes.NewBufferString(string(js))
	resp, err := http.Post(worker.Address+routes.BigCodeApplication.GetSuggestion(mpNumString), "text/json", b)
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		revel.INFO.Println(err)
		return res, err
	}
	revel.INFO.Println(string(body))
	if err := json.Unmarshal(body, &res); err != nil {
		revel.INFO.Println(err)
		return res, err
	}
	return res, err
}

func (c AdminApplication) ShowAllStudentsMachineProblems(mpNumString string) revel.Result {

	mpNum, err := strconv.Atoi(mpNumString)
	if err != nil {
		c.Flash.Error("Invalid mp Id")
		return c.Render(routes.PublicApplication.Index())
	}

	admin := c.connected()
	if !models.IsAdmin(admin) {
		c.Flash.Error("No admin privilages")
		return c.Render(routes.PublicApplication.Index())
	}

	conf, _ := ReadMachineProblemConfig(mpNum)

	students, err := models.FindUsersByMachineProblemNumber(mpNum)
	if err != nil {
		c.Flash.Error("Invalid mp number")
		return c.Render(routes.PublicApplication.Index())
	}

	class_roster := make([]studentRecord, len(students))
	for idx, student := range students {
		mp, _ := models.FindMachineProblemByUser(student, mpNum)
		grade, _ := models.FindGradeByMachineProblem(mp)
		attempt, _ := models.FindAttempt(grade.AttemptInstanceId)
		program, _ := models.FindProgram(attempt.ProgramInstanceId)
		qs, _ := models.FindQuestionsByMachineProblem(mp)
		qis, _ := models.FindQuestionItems(mp, qs)

		class_roster[idx] = studentRecord{
			Id:             student.Id,
			Student:        student,
			MachineProblem: mp,
			Questions:      qis,
			Grade:          grade,
			Program:        program,
			Attempt:        attempt,
		}
	}
	c.RenderArgs["admin"] = admin
	c.RenderArgs["mp_num"] = mpNum
	c.RenderArgs["mp_config"] = conf
	c.RenderArgs["class_roster"] = class_roster

	return c.Render()
}

func (c AdminApplication) ExportMachineProblemsGradesToCSV(mpNumString string) revel.Result {

	mpNum, err := strconv.Atoi(mpNumString)
	if err != nil {
		c.Flash.Error("Invalid mp Id")
		return c.Render(routes.PublicApplication.Index())
	}

	user := c.connected()
	if !models.IsAdmin(user) {
		c.Flash.Error("No admin privilages")
		return c.Render(routes.PublicApplication.Index())
	}

	students, err := models.FindUsersByMachineProblemNumber(mpNum)
	if err != nil {
		c.Flash.Error("Invalid mp number")
		return c.Render(routes.PublicApplication.Index())
	}

	csv := make([]string, len(students)+1)
	csv[0] = "username,email,mp_num,code_grade,questions_grade,total_grade,updated"
	for idx, student := range students {
		mp, _ := models.FindMachineProblemByUser(student, mpNum)
		grade, _ := models.FindGradeByMachineProblem(mp)
		csv[idx+1] = strings.Join([]string{
			student.UserName,
			student.Email,
			strconv.Itoa(mpNum),
			strconv.Itoa(int(grade.CodeScore)),
			strconv.Itoa(int(grade.PeerReviewScore)),
			strconv.Itoa(int(grade.TotalScore)),
			grade.Updated.Format(time.RFC3339),
		}, ",")
	}

	return c.RenderText(strings.Join(csv, "\n"))
}

func sendEmailToStudent(student models.User) {

	templt := template.Must(template.New("taEmailTemplate").Parse(taEmailTemplate))

	var buf bytes.Buffer

	if err := templt.Execute(&buf, student); err != nil {
		revel.TRACE.Println(err)
		return
	}
	SendEmail(ApplicationEmail, student.Email, "TA has graded your MP", string(buf.Bytes()))
}

func (c AdminApplication) UpdateStudentMachineProblem(studentIdString string, mpNumString string) revel.Result {
	type gradeInformation struct {
		CodeScore           int64  `json:"code_score"`
		CodeInspectionScore int64  `json:"code_inspection_score"`
		CodeComment         string `json:"code_comment"`
		QuestionsScore      int64  `json:"questions_score"`
		QuestionsComment    string `json:"questions_comment"`
		TaCodeText          string `json:"ta_code_text"`
	}
	var gradeUpdate gradeInformation
	var gradeString string

	studentId, err := strconv.Atoi(studentIdString)
	if err != nil {
		c.Flash.Error("Invalid user Id")
		return c.Render(routes.PublicApplication.Index())
	}
	mpNum, err := strconv.Atoi(mpNumString)
	if err != nil {
		c.Flash.Error("Invalid mp Id")
		return c.Render(routes.PublicApplication.Index())
	}

	student, err := models.FindUser(int64(studentId))
	if err != nil {
		return c.RenderJson(map[string]interface{}{
			"status": "error",
			"title":  "No user found",
			"data":   "Cannot find user.",
			"error":  err,
		})
	}

	admin := c.connected()
	if !models.IsAdmin(admin) {
		return c.RenderJson(map[string]interface{}{
			"status": "error",
			"title":  "No admin privilages",
			"data":   "Cannot perform request because of lack of admin privilages.",
			"error":  "",
		})
	}

	c.Params.Bind(&gradeString, "grade")
	if err := json.Unmarshal([]byte(gradeString), &gradeUpdate); err != nil {
		return c.RenderJson(map[string]interface{}{
			"status": "error",
			"title":  "Error: Cannot parse grade update request.",
			"data":   "The system was not able to parse your grade update request.",
			"error":  err,
		})
	}

	mp, _ := models.FindMachineProblemByUser(student, mpNum)

	if _, err := models.UpdateGradeTA(mp, gradeUpdate.CodeScore, gradeUpdate.CodeInspectionScore,
		gradeUpdate.QuestionsScore, gradeUpdate.CodeComment, gradeUpdate.QuestionsComment,
		gradeUpdate.TaCodeText); err != nil {
		return c.RenderJson(map[string]interface{}{
			"status": "error",
			"title":  "Not able to update the database",
			"data":   "Cannot update the grades database.",
			"error":  err,
		})
	}
	//sendEmailToStudent(student)
	link := "/admin/student/" + strconv.Itoa(int(student.Id)) + "/mp/" + strconv.Itoa(int(mp.Number))
	return c.RenderJson(map[string]interface{}{
		"status": "success",
		"title":  "Grade has been updated",
		"link":   link,
	})
}

func InitAdminController() {

}
