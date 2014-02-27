package models

import (
	"errors"
	"time"
	. "wb/app/config"
)

//
type Attempt struct {
	Id                       int64     `json:"id" qbs:"pk,notnull"`
	MachineProblemInstanceId int64     `json:"-" qbs:"fk:MachineProblem"`
	ProgramInstanceId        int64     `json:"-" qbs:"fk:Program"`
	DatasetId                int       `json:"dataset_id" qbs:"default:'0'"`
	CompilationFailed        bool      `json:"compilation_failed" qbs:"default:'0'"`
	CompileStdout            string    `json:"compile_stdout" qbs:"default:''"`
	CompileStderr            string    `json:"compile_stderr" qbs:"default:''"`
	TimeoutError             bool      `json:"timeout_error"`
	TimeoutValue             float64   `json:"timeout_value"`
	RunFailed                bool      `json:"run_failed" qbs:"default:'0'"`
	RunStdout                string    `json:"run_stdout" qbs:"default:''"`
	RunStderr                string    `json:"run_stderr" qbs:"default:''"`
	Sandboxed                bool      `json:"sandboxed" qbs:"default:'0'"`
	SandboxKeyword           string    `json:"sandbox_keyword" qbs:"default:''"`
	CompileElapsedTime       int64     `json:"compile_elapsed_time" qbs:"default:'0'"`
	RunElapsedTime           int64     `json:"run_elapsed_time" qbs:"default:'0'"`
	CompileStartTime         time.Time `json:"-"`
	CompileEndTime           time.Time `json:"-"`
	RunStartTime             time.Time `json:"-"`
	RunEndTime               time.Time `json:"-"`
	RequestStartTime         time.Time `json:"-"`
	RequestEndTime           time.Time `json:"-"`
	SolutionCorrect          bool      `json:"solution_correct" qbs:"default:'0'"`
	SolutionMessage          string    `json:"solution_message"`
	UserOutput               string    `json:"user_output"`
	Language                 string    `json:"language"`
	RunId                    string    `json:"run_id"`
	OnAllDatasets            bool      `json:"-" qbs:"default:'0'"`
	InternalCData            string    `json:"internal_c_data"`
	GradedQ                  bool      `json:"graded"`
	Reserved1                string    `json:"-"`
	Reserved2                string    `json:"-"`
	Reserved3                string    `json:"-"`
	Reserved4                string    `json:"-"`
	Reserved5                string    `json:"-"`
	Updated                  time.Time `json:"updated"`
	Created                  time.Time `json:"created"`
}

// Create attempt table if it already does not exist (Database Migration)
func CreateAttemptTable() error {
	/*
		migration, err := qbs.GetMigration()
		if err != nil {
			stats.ERROR.Println("Failed to attempt database:", err)
			return err
		} else {
			stats.TRACE.Println("Created attempt table")
		}
		defer migration.Close()
		return migration.CreateTableIfNotExists(new(Attempt))
	*/
	return nil
}

func FindAttempt(id int64) (Attempt, error) {
	var attempt Attempt

	if err := DB.First(&attempt, id).Error; err != nil {
		return attempt, err
	}
	return attempt, nil
}

func FindAttemptByRunId(id string) (Attempt, error) {
	var attempt Attempt
	err := DB.
		Where("run_id = ?", id).
		Order("id DESC").
		First(&attempt).
		Error
	return attempt, err
}

func FindAllAttemptsByRunId(id string) ([]Attempt, error) {
	var attempts []Attempt
	err := DB.
		Where("run_id = ?", id).
		Order("dataset_id").
		Order("id DESC").
		Find(&attempts).
		Error
	if err != nil {
		return attempts, err
	} else if len(attempts) == 0 {
		return attempts, errors.New("Cannot find any attempts")
	}
	return attempts, err
}

func FindAttemptsByMachineProblem(mp MachineProblem) ([]Attempt, error) {
	var attempts []Attempt
	err := DB.
		Where("machine_problem_instance_id = ?", mp.Id).
		Order("id DESC").
		Find(&attempts).
		Error
	if err != nil {
		return nil, err
	} else if len(attempts) == 0 {
		return nil, errors.New("Cannot find any attempts")
	}
	return attempts, nil
}

func FindLastAttempt() (Attempt, error) {
	var attempt Attempt
	err := DB.
		Order("id DESC").
		First(&attempt).
		Error
	if err != nil {
		return attempt, err
	}

	return attempt, nil
}

func FindLastAttemptByMachineProblem(mp MachineProblem) (Attempt, error) {
	var attempt Attempt
	err := DB.
		Where("machine_problem_instance_id = ?", mp.Id).
		Order("id DESC").
		First(&attempt).
		Error
	if err != nil {
		return attempt, err
	}
	return attempt, nil
}

func CreateDummyAttemptByMachineProblem(mp MachineProblem) (Attempt, error) {
	prog, err := RecentProgram(mp)
	if err != nil {
		conf, _ := ReadMachineProblemConfig(mp.Number)
		if prog, err = CreateProgram(mp, conf.CodeTemplate); err != nil {
			return Attempt{}, err
		}
	}
	attempt := Attempt{
		MachineProblemInstanceId: mp.Id,
		ProgramInstanceId:        prog.Id,
		RunFailed:                true,
		RunStdout:                "No attempt found",
		Created:                  time.Now(),
		Updated:                  time.Now(),
	}
	err = DB.Save(&attempt).Error
	return attempt, err
}
