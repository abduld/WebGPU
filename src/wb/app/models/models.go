package models

import (
	. "wb/app/config"
	"wb/app/stats"

	"github.com/abduld/gorm"
	_ "github.com/go-sql-driver/mysql"
	"github.com/robfig/revel"
)

var (
	DB gorm.DB
)

func dbRegister() {
	var err error
	if err != nil {
		stats.ERROR.Println("Cannot connect to database: ", err)
	}

	if db, err := gorm.Open(DatabaseProvider, DatabaseSourceName); err == nil {
		DB = db

		DB.SingularTable(true)
		//DB.LogMode(true)
		DB.SetLogger(gorm.Logger{revel.INFO})
		DB.DB().SetMaxIdleConns(1024)
		DB.DB().SetMaxOpenConns(4098)
		DB.DB().Ping()
	}
}

func dbMigrate() {
	CreateUserTable()
	CreateMachineProblemTable()
	CreateProgramTable()
	CreateAttemptTable()
	CreateGradeTable()
	CreatePeerReviewTable()
	CreateQuestionsTable()
}

func InitModels() {
	if IsMaster {
		dbRegister()
		dbMigrate()
	}
}

func ResetModels() {
}
