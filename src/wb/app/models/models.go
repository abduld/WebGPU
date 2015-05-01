package models

import (
	"time"
	. "wb/app/config"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
	"github.com/revel/revel"
)

var (
	DB gorm.DB
)

func updateTimeStampWhenCreate(scope *gorm.Scope) {
	if !scope.HasError() {
		now := time.Now()
		scope.SetColumn("Created", now)
		scope.SetColumn("Updated", now)
	}
}

func updateTimeStampWhenUpdate(scope *gorm.Scope) {
	if !scope.HasError() {
		now := time.Now()
		scope.SetColumn("Updated", now)
	}
}

func dbRegister() {

	if db, err := gorm.Open(DatabaseProvider, DatabaseSourceName); err == nil {
		DB = db

		DB.SingularTable(true)
		DB.LogMode(false)
		//DB.SetLogger(gorm.Logger{revel.INFO})
		DB.Callback().Create().Replace("gorm:update_time_stamp_when_create", updateTimeStampWhenCreate)
		DB.Callback().Create().Replace("gorm:update_time_stamp_when_update", updateTimeStampWhenUpdate)

		DB.DB().Ping()
		DB.DB().SetMaxIdleConns(256)
		DB.DB().SetMaxOpenConns(512)

	} else {
		revel.ERROR.Println("Cannot start database")
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
	CreateBigcodeVoteTable()
}

func InitModels() {
	if IsMaster {
		dbRegister()
		dbMigrate()
	}
}

func ResetModels() {
}
