package models

import (
	"errors"
	"time"
	. "wb/app/config"
)

type Attempt struct {
	Id                       int64     `json:"id" gorm:"column:id; primary_key:yes;"`
	MachineProblemInstanceId int64     `json:"-" gorm:"column:machine_problem_instance_id" sql:"not null"`
	ProgramInstanceId        int64     `json:"-" gorm:"column:program_instance_id" sql:"not null"`
	DatasetId                int       `json:"dataset_id" gorm:"column:dataset_id" sql:"default:'0'"`
	CompilationFailed        bool      `json:"compilation_failed" gorm:"column:compilation_failed" sql:"default:'0'"`
	CompileStdout            string    `json:"compile_stdout" gorm:"column:compile_stdout" sql:"default:''"`
	CompileStderr            string    `json:"compile_stderr" gorm:"column:compile_stderr" sql:"default:''"`
	TimeoutError             bool      `json:"timeout_error" gorm:"column:timeout_error"`
	TimeoutValue             float64   `json:"timeout_value" gorm:"column:timeout_value"`
	RunFailed                bool      `json:"run_failed" gorm:"column:run_failed" sql:"default:'0'"`
	RunStdout                string    `json:"run_stdout" gorm:"column:run_stdout" sql:"default:''"`
	RunStderr                string    `json:"run_stderr" gorm:"column:run_stderr" sql:"default:''"`
	Sandboxed                bool      `json:"sandboxed" gorm:"column:sandboxed" sql:"default:'0'"`
	SandboxKeyword           string    `json:"sandbox_keyword" gorm:"column:sandbox_keyword" sql:"default:''"`
	CompileElapsedTime       int64     `json:"compile_elapsed_time" gorm:"column:compile_elapsed_time" sql:"default:'0'"`
	RunElapsedTime           int64     `json:"run_elapsed_time" gorm:"column:run_elapsed_time" sql:"default:'0'"`
	CompileStartTime         time.Time `json:"-" gorm:"column:compile_start_time"`
	CompileEndTime           time.Time `json:"-" gorm:"column:compile_end_time"`
	RunStartTime             time.Time `json:"-" gorm:"column:run_start_time"`
	RunEndTime               time.Time `json:"-" gorm:"column:run_end_time"`
	RequestStartTime         time.Time `json:"-" gorm:"column:request_start_time"`
	RequestEndTime           time.Time `json:"-" gorm:"column:request_end_time"`
	SolutionCorrect          bool      `json:"solution_correct" gorm:"column:solution_correct" sql:"default:'0'"`
	SolutionMessage          string    `json:"solution_message" gorm:"column:solution_message"`
	UserOutput               string    `json:"user_output" gorm:"column:user_output"`
	Language                 string    `json:"language" gorm:"column:language"`
	RunId                    string    `json:"run_id" gorm:"column:run_id"`
	OnAllDatasets            bool      `json:"-"  gorm:"column:on_all_datasets" sql:"default:'0'"`
	InternalCData            string    `json:"internal_c_data" gorm:"column:internal_c_data"`
	GradedQ                  bool      `json:"graded" gorm:"column:graded_q"`
	Reserved1                string    `json:"-" gorm:"column:reserved1"`
	Reserved2                string    `json:"-" gorm:"column:reserved2"`
	Reserved3                string    `json:"-" gorm:"column:reserved3"`
	Reserved4                string    `json:"-" gorm:"column:reserved4"`
	Reserved5                string    `json:"-" gorm:"column:reserved5"`
	Updated                  time.Time `json:"updated" gorm:"column:updated"`
	Created                  time.Time `json:"created" gorm:"column:created"`
}

// Create attempt table if it already does not exist (Database Migration)
func CreateAttemptTable() error {
	/*
		migration, err := sql.GetMigration()
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
		Order("id DESC").
		First(&attempt, Attempt{RunId: id}).
		Error
	return attempt, err
}

func FindAllAttemptsByRunId(id string) ([]Attempt, error) {
	var attempts []Attempt
	err := DB.
		Order("dataset_id").
		Order("id DESC").
		Find(&attempts, Attempt{RunId: id}).
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
		Order("id DESC").
		Find(&attempts, Attempt{MachineProblemInstanceId: mp.Id}).
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
		Order("id DESC").
		First(&attempt, Attempt{MachineProblemInstanceId: mp.Id}).
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
