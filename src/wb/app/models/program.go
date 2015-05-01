package models

import "time"

//
type Program struct {
	Id                       int64     `json:"id" gorm:"column:id; primary_key:yes"`
	MachineProblemInstanceId int64     `json:"-" gorm:"column:machine_problem_instance_id" sql:"not null"`
	Language                 string    `json:"language" gorm:"column:language" sql:"not null"`
	Text                     string    `json:"program_text" gorm:"column:text" sql:"default:''"`
	Reserved1                string    `json:"-" gorm:"column:reserved1"`
	Reserved2                string    `json:"-" gorm:"column:reserved2"`
	Reserved3                string    `json:"-" gorm:"column:reserved3"`
	Reserved4                string    `json:"-" gorm:"column:reserved4"`
	Reserved5                string    `json:"-" gorm:"column:reserved5"`
	Created                  time.Time `json:"created" gorm:"column:created"`
	BigCodeText         string    `json:"bigcode_text" gorm:"column:big_code_text" sql:"default:''"`
	BigCodeSignature         string    `json:"bigcode_signature" gorm:"column:bigcode_signature" sql:"default:''"`
	Hidden                   bool      `json:"hidden" gorm:"column:hidden" sql:"default:'0'"`
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
		Hidden:  false,
		Created: time.Now(),
	}
	err := DB.Save(&prog).Error
	return prog, err
}

// Finds all the programs in the database that are associated
// with a machine problem id
func FindProgramsByMachineProblem(mp MachineProblem) ([]Program, error) {
	var progs []Program
	err := DB.
		Order("id DESC").
		Find(&progs, Program{MachineProblemInstanceId: mp.Id, Hidden: false}).
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
		Order("id DESC").
		First(&prog, Program{MachineProblemInstanceId: mp.Id, Hidden: false}).
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
		Order("id DESC").
		Limit(count).
		Find(&progs, Program{MachineProblemInstanceId: mp.Id, Hidden: false}).
		Error
	return progs, err
}

// Gets all the correct programs for a given MP
func CorrectProgramsForMP(mpNumber int, limit int) ([]Program, error) {
	var progs []Program

	err := DB.
		Joins("inner JOIN attempt ON attempt.program_instance_id = program.id "+
		"inner JOIN grade ON grade.attempt_instance_id = attempt.id "+
		"inner JOIN machine_problem ON machine_problem.id = grade.machine_problem_id").
		Where("machine_problem.number = ? and grade.total_score = ?", mpNumber, 100).
		Select("program.*").
		Limit(limit).
		Find(&progs).
		Error

	return progs, err
}

// Gets all the correct programs for a given MP
func AllCorrectProgramsForMP(mpNumber int) ([]Program, error) {
	var progs []Program

	err := DB.
		Joins("inner JOIN attempt ON attempt.program_instance_id = program.id "+
		"inner JOIN grade ON grade.attempt_instance_id = attempt.id "+
		"inner JOIN machine_problem ON machine_problem.id = grade.machine_problem_id").
		Where("machine_problem.number = ? and grade.total_score = ?", mpNumber, 100).
		Select("program.*").
		Find(&progs).
		Error

	return progs, err
}
