package models

import (
	"errors"
	"math"
	"strconv"
	"strings"
	"time"
	. "wb/app/config"
)

type Grade struct {
	Id                         int64     `json:"id" gorm:"column:id; primary_key:yes"`
	AttemptInstanceId          int64     `json:"attempt_id" gorm:"column:attempt_instance_id" sql:"not null"`
	RunId                      string    `json:"run_id" gorm:"column:run_id"`
	MachineProblemId           int64     `json:"mp_id" gorm:"column:machine_problem_id" sql:"not null"`
	DatasetsScored             string    `json:"datasets_scored" gorm:"column:datasets_scored"`
	PeerReviewPercentage       float32   `json:"peer_review_percentage" gorm:"column:peer_review_percentage"`
	PeerReviewScore            int64     `json:"peer_review_score" gorm:"column:peer_review_score" sql:"default:'0'"`
	CodePercentage             float32   `json:"code_percentage" gorm:"column:code_percentage"`
	CodeScore                  int64     `json:"code_score" gorm:"column:code_score" sql:"default:'0'"`
	TotalScore                 int64     `json:"total_score" gorm:"column:total_score" sql:"default:'0'"`
	Text                       string    `json:"text" gorm:"column:text" sql:"default:''"`
	Completed                  bool      `json:"completed" gorm:"column:completed" sql:"default:'0'"`
	Reasons                    string    `json:"reasons" gorm:"column:reasons"`
	CourseraSynced             bool      `json:"coursera_synced" gorm:"column:coursera_synced"`
	CourseraCodingGrade        int64     `json:"coursera_coding_grade" gorm:"column:coursera_coding_grade"`
	CourseraPeerReviewGrade    int64     `json:"coursera_peer_review_grade" gorm:"column:coursera_peer_review_grade"`
	CourseraGrade              int64     `json:"coursera_grade" gorm:"column:coursera_grade"`
	Reserved1                  string    `json:"-" gorm:"column:reserved1"`
	Reserved2                  string    `json:"-" gorm:"column:reserved2"`
	Reserved3                  string    `json:"-" gorm:"column:reserved3"`
	Reserved4                  string    `json:"-" gorm:"column:reserved4"`
	Reserved5                  string    `json:"-" gorm:"column:reserved5"`
	Updated                    time.Time `json:"updated" gorm:"column:updated"`
	Created                    time.Time `json:"created" gorm:"column:created"`
	TaGraded                   bool      `json:"ta_graded" gorm:"column:ta_graded"`
	TaCodeComment              string    `json:"ta_code_comment" gorm:"column:ta_code_comment"`
	TaQuestionsComment         string    `json:"ta_code_comment" gorm:"column:ta_questions_comment"`
	TaCodeInspectionScore      int64     `json:"ta_code_inspection_score" gorm:"column:ta_code_inspection_score"`
	TaCodeInspectionPercentage float32   `json:"ta_code_inspection_percentage" gorm:"column:ta_code_inspection_percentage"`
	QuestionsScore             int64     `json:"questions_score" gorm:"column:questions_score"`
	TaCodeText                 string    `json:"ta_code_text" gorm:"column:ta_code_text"`
}

// Create grade table if it already does not exist (Database Migration)
func CreateGradeTable() error {
	/*
		migration, err := qbs.GetMigration()
		if err != nil {
			stats.ERROR.Println("Failed to migrate grade database:", err)
			return err
		} else {
			stats.TRACE.Println("Created grade table")
		}
		defer migration.Close()
		return migration.CreateTableIfNotExists(new(Grade))
	*/
	return nil
}

func FindGrade(id int64) (Grade, error) {
	var grade Grade
	err := DB.First(&grade, id).Error
	if err != nil {
		return grade, err
	}
	return grade, nil
}

func FindGrades() ([]Grade, error) {
	var grades []Grade
	err := DB.
		Order("updated DESC").
		Find(&grades).
		Error

	if err != nil {
		return nil, err
	} else if len(grades) == 0 {
		return nil, errors.New("Could not find any grade")
	}
	return grades, nil
}

func FindGradesByAttempt(attempt Attempt) ([]Grade, error) {
	var grades []Grade
	err := DB.
		Order("updated DESC").
		Find(&grades, Grade{AttemptInstanceId: attempt.Id}).
		Error
	if err != nil {
		return nil, err
	} else if len(grades) == 0 {
		return nil, errors.New("Cannot find grade.")
	}
	return grades, err
}

func FindGradeByAttempt(attempt Attempt) (Grade, error) {
	var grade Grade
	err := DB.
		Order("id DESC").
		First(&grade, Grade{AttemptInstanceId: attempt.Id}).
		Error
	if err != nil {
		return grade, err
	}
	return grade, err
}

func FindGradeByRunId(runId string) (Grade, error) {
	var grade Grade
	err := DB.
		Order("id DESC").
		First(&grade, Grade{RunId: runId}).
		Error

	if err != nil {
		return grade, err
	}
	return grade, nil
}

func FindGradesByMachineProblem(mp MachineProblem) ([]Grade, error) {

	var grades []Grade
	err := DB.
		Order("id DESC").
		Find(&grades, Grade{MachineProblemId: mp.Id}).
		Error

	if err != nil {
		return nil, err
	}
	return grades, nil
}

func FindAllGradesByMachineProblem(mp MachineProblem) ([]Grade, error) {
	grades, err := FindGradesByMachineProblem(mp)
	if err != nil {
		return nil, err
	} else if len(grades) == 0 {
		return nil, errors.New("Cannot find grade")
	}
	gs := []Grade{}
	for _, g := range grades {
		if g.Id > 0 {
			gs = append(gs, g)
		}
	}
	if len(gs) == 0 {
		return nil, errors.New("Cannot find grade")
	}
	return gs, nil
}

func FindGradeByMachineProblem(mp MachineProblem) (Grade, error) {
	var grade Grade
	err := DB.
		Order("id DESC").
		First(&grade, Grade{MachineProblemId: mp.Id}).
		Error
	return grade, err
}

func AllGraded(grade Grade) bool {
	//revel.TRACE.Println(grade.MachineProblemId)
	mp, err := FindMachineProblem(grade.MachineProblemId)
	if err != nil {
		//revel.TRACE.Println("Cannot find machine problem...")
		return false
	}

	mpConfig, _ := ReadMachineProblemConfig(mp.Number)

	if !strings.Contains(grade.DatasetsScored, "(compile)") {
		return false
	}

	for _, dataset := range mpConfig.Datasets {
		s := strconv.Itoa(dataset.Id)
		if !strings.Contains(grade.DatasetsScored, "("+s+")") {
			return false
		}
	}
	return true
}

func CopyGrade(old Grade) (Grade, error) {
	grade := old
	grade.Id = 0

	return grade, nil
}

func CreateGrade(attempt Attempt) (Grade, error) {
	var grade Grade

	grade.AttemptInstanceId = attempt.Id
	grade.MachineProblemId = attempt.MachineProblemInstanceId

	mp, err := FindMachineProblem(attempt.MachineProblemInstanceId)
	if err != nil {
		return grade, err
	}

	conf, _ := ReadMachineProblemConfig(mp.Number)

	grade.PeerReviewPercentage = conf.PeerReviewScore
	grade.CodePercentage = conf.CodeScore
	grade.Created = time.Now()

	return CopyGrade(grade)
}

func keywordToDescription(keyword string) string {
	switch keyword {
	case "__shared__":
		return "use shared memory"
	case "__syncthreads":
		return "synchronize the threads"
	case "syncthreads":
		return "synchronize the threads"
	case "atomicAdd":
		return "use atomic operations"
	default:
	}
	return "use the " + keyword + " keyword"
}

func fromPercentage32(val float32) int64 {
	return int64(math.Floor(float64(100.0)*float64(val) + float64(0.5)))
}

func fromPercentage64(val float64) int64 {
	return int64(math.Floor(float64(100.0)*float64(val) + float64(0.5)))
}

func UpdateGradeWithAttempts(attempts []Attempt) (Grade, error) {

	if len(attempts) == 0 {
		return Grade{}, errors.New("Invalid attempts")
	}

	firstAttempt := attempts[0]

	mp, err := FindMachineProblem(firstAttempt.MachineProblemInstanceId)
	if err != nil {
		return Grade{}, err
	}

	g, err := FindGradeByMachineProblem(mp)

	var grade Grade
	if err != nil {
		//revel.TRACE.Println("Creating grade..")
		grade, err = CreateGrade(firstAttempt)
	} else {
		//revel.TRACE.Println("copy grade..")
		grade, err = CopyGrade(g)
	}

	if err != nil {
		return grade, errors.New("Cannot create grade....")
	}

	grade.CodeScore = 0
	grade.DatasetsScored = ""
	grade.Reasons = ""
	grade.RunId = firstAttempt.RunId
	grade.MachineProblemId = firstAttempt.MachineProblemInstanceId

	grade.AttemptInstanceId = firstAttempt.Id

	grade.Created = time.Now()
	grade.Updated = time.Now()

	mpNum := mp.Number
	mpConfig, _ := ReadMachineProblemConfig(mpNum)

	for ii, attempt := range attempts {
		if attempt.Id == 0 {
			continue
		}
		if ii == 0 {
			grade.DatasetsScored += ",(compile)"

			score := fromPercentage32(mpConfig.CompileScore)

			if attempt.CompilationFailed && score > 0 {
				scoreStr := strconv.Itoa(int(score))
				if !strings.HasSuffix(grade.Reasons, ",") {
					grade.Reasons += ","
				}
				grade.Reasons += "Lost " + scoreStr + " points in the coding section because program failed to compile."
			} else {
				grade.CodeScore += score
			}

			if !strings.Contains(grade.DatasetsScored, "(keywords)") {
				prog, _ := FindProgram(attempt.ProgramInstanceId)
				for _, keyword := range mpConfig.Keywords {
					score := fromPercentage32(keyword.Score)
					if strings.Contains(prog.Text, keyword.Data) {
						grade.CodeScore += score
					} else if score >= 0 {
						scoreStr := strconv.Itoa(int(score))
						if !strings.HasSuffix(grade.Reasons, ",") {
							grade.Reasons += ","
						}
						grade.Reasons += "Lost " + scoreStr + " points in the coding section because program did not " +
							keywordToDescription(keyword.Data) + "."
					}
				}
				grade.DatasetsScored += ",(keywords)"
			}
		}

		if attempt.DatasetId >= 0 {
			datasetIdStr := strconv.Itoa(attempt.DatasetId)
			grade.DatasetsScored += ",(" + datasetIdStr + ")"
			score := fromPercentage64(mpConfig.Datasets[attempt.DatasetId].Score)

			if attempt.SolutionCorrect {
				grade.CodeScore += score
			} else {
				scoreStr := strconv.Itoa(int(score))
				if !strings.HasSuffix(grade.Reasons, ",") {
					grade.Reasons += ","
				}
				grade.Reasons += "Lost " + scoreStr + " points in the coding section because program did not produce " +
					"correct answer for dataset " + datasetIdStr + "."
			}
		}
	}

	grade.CodeScore = int64(math.Ceil(math.Min(float64(grade.CodeScore), float64(100*mpConfig.CodeScore))))
	grade.TotalScore += grade.CodeScore
	grade.TotalScore = int64(math.Min(float64(grade.TotalScore), float64(100)))
	grade.Updated = time.Now()

	err = DB.Save(&grade).Error
	return grade, err
}

func UpdateGradePeerReview(user User, mp MachineProblem) (Grade, error) {
	grade, err := FindGradeByMachineProblem(mp)

	if err != nil {
		lastAttempt, err := FindLastAttemptByMachineProblem(mp)
		if err != nil {
			return grade, errors.New("No attempts for this MP were found")
		}
		grade, err = CreateGrade(lastAttempt)
		if err != nil {
			return grade, errors.New("Cannot create a new grade.")
		}
	}

	grade, err = CopyGrade(grade)
	if err != nil {
		return grade, errors.New("Cannot create grade.")
	}

	prs, err := GetPeerReviewsByReviewerAndMachineProblem(user, mp)
	if err != nil {
		return grade, err
	}

	mpConfig, _ := ReadMachineProblemConfig(mp.Number)

	score := fromPercentage32(mpConfig.PeerReviewScore)
	scoreStr := strconv.Itoa(int(score))
	reasonString := ",Lost " + scoreStr + " points in the peer review section because of an incomplete peer review."

	for _, pr := range prs {
		if strings.TrimSpace(pr.QuestionReviewComment) == "" ||
			strings.TrimSpace(pr.CodeReviewComment) == "" {
			if !strings.Contains(grade.Reasons, reasonString) {
				grade.Reasons += reasonString
			}
			err = DB.Save(&grade).Error
			return grade, err
		}
	}

	grade.PeerReviewScore = int64(math.Ceil(math.Min(float64(score),
		float64(100*mpConfig.PeerReviewScore))))

	if score >= 0 && strings.Contains(grade.Reasons, reasonString) {
		grade.Reasons = strings.Replace(grade.Reasons, reasonString, "", -1)
	}

	if !strings.Contains(grade.DatasetsScored, "(peer_review)") {
		grade.TotalScore += int64(math.Ceil(math.Min(float64(score), 100)))
		grade.DatasetsScored += ",(peer_review)"
	}

	grade.Updated = time.Now()

	err = DB.Save(&grade).Error
	return grade, err
}

func UpdateGradeTA(mp MachineProblem, codeGrade int64, inspectionGrade int64, questionsGrade int64, codeComment string, questionsComment string, taCodeText string) (Grade, error) {
	grade, err := FindGradeByMachineProblem(mp)

	if err != nil {
		lastAttempt, err := FindLastAttemptByMachineProblem(mp)
		if err != nil {
			return grade, errors.New("No attempts for this MP were found")
		}
		grade, err = CreateGrade(lastAttempt)
		if err != nil {
			return grade, errors.New("Cannot create a new grade.")
		}
	} else {
		grade, err = CopyGrade(grade)
	}

	if err != nil {
		return grade, errors.New("Cannot create grade.")
	}

	grade.CodeScore = codeGrade
	grade.QuestionsScore = questionsGrade
	grade.TaGraded = true
	grade.TaCodeComment = codeComment
	grade.TaQuestionsComment = questionsComment
	grade.TaCodeInspectionScore = inspectionGrade
	grade.TotalScore = codeGrade + questionsGrade + inspectionGrade
	grade.TaCodeText = taCodeText
	if grade.TotalScore > 100 {
		grade.TotalScore = 100
	}

	if !strings.Contains(grade.DatasetsScored, "(ta)") {
		grade.DatasetsScored += ",(ta)"
	}

	grade.Updated = time.Now()

	err = DB.Save(&grade).Error
	return grade, err
}

func UpdateCourseraGrade(grade Grade, kind string, score int64) (Grade, error) {
	grade.Id = 0

	if kind == "code" {
		grade.CourseraCodingGrade = score
	} else {
		grade.CourseraPeerReviewGrade = score
	}

	grade.Updated = time.Now()

	err := DB.Save(&grade).Error
	return grade, err
}

func RandomGradeByMachineProblem(mpNumber int) (Grade, error) {
	/*
	   SELECT grade.* FROM wb.grade
	   INNER JOIN wb.machine_problem mp ON grade.machine_problem_id = mp.id
	   WHERE mp.number = mpNumber;
	*/
	grade := Grade{}
	err := DB.
		Select("grade.id").
		Joins("inner join machine_problem mp on grade.machine_problem_id = mp.id").
		Where("mp.number = ?", mpNumber).
		Order("RAND()").
		Select("grade.*").
		Limit(1).
		First(&grade).
		Error
	return grade, err
}
