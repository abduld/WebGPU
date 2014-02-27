package models

import "time"

//
type Program struct {
	Id                       int64     `json:"id" qbs:"pk,notnull"`
	MachineProblemInstanceId int64     `json:"-" qbs:"fk:MachineProblem"`
	Language                 string    `json:"language"`
	Text                     string    `json:"program_text" qbs:"default:''"`
	Reserved1                string    `json:"-"`
	Reserved2                string    `json:"-"`
	Reserved3                string    `json:"-"`
	Reserved4                string    `json:"-"`
	Reserved5                string    `json:"-"`
	Created                  time.Time `json:"created"`
}

// Create program table if it already does not exist (Database Migration)
func CreateProgramTable() error {
	/*
		migration, err := qbs.GetMigration()
		if err != nil {
			stats.ERROR.Println("Failed to migrate program database:", err)
			return err
		} else {
			stats.TRACE.Println("Created program table")
		}
		defer migration.Close()
		return migration.CreateTableIfNotExists(new(Program))
	*/
	return nil
}

// Creates a new program in the database using
// the program text and mp key as input
func CreateProgram(mp MachineProblem, text string) (Program, error) {
	prog := Program{
		MachineProblemInstanceId: mp.Id,
		Text:    text,
		Created: time.Now(),
	}
	err := DB.Save(&prog).Error
	return prog, err
}

// Finds all the programs in the database that are associated
// with a machine problem id
func FindPrograms(mp MachineProblem) ([]Program, error) {
	var progs []Program
	err := DB.
		Where("machine_problem_instance_id = ?", mp.Id).
		Find(&progs).
		Error
	return progs, err
}

// Finds a program with specified pk
func FindProgram(id int64) (Program, error) {
	var prog Program
	err := DB.First(&prog, id).Error
	if err != nil {
		return prog, err
	}
	return prog, err
}

// Gets the most recent program
func RecentProgram(mp MachineProblem) (Program, error) {
	var prog Program
	err := DB.
		Where("machine_problem_instance_id = ?", mp.Id).
		Order("id DESC").
		First(&prog).
		Error
	if err != nil {
		return prog, err
	}
	return prog, err
}

// Gets the most recent program
func RecentPrograms(mp MachineProblem, count int) ([]Program, error) {
	var progs []Program
	err := DB.
		Where("machine_problem_instance_id = ?", mp.Id).
		Order("id DESC").
		Limit(count).
		Find(&progs).
		Error
	return progs, err
}
