package models

import (
	"time"
)

//
type MachineProblem struct {
	Id             int64     `json:"id" gorm:"column:id; primary_key:yes"`
	Number         int       `json:"number" sql:"not null"`
	UserInstanceId int64     `json:"user_id" gorm:"column:user_instance_id" sql:"not null"`
	Reserved1      string    `json:"-" gorm:"column:reserved1"`
	Reserved2      string    `json:"-" gorm:"column:reserved2"`
	Reserved3      string    `json:"-" gorm:"column:reserved3"`
	Reserved4      string    `json:"-" gorm:"column:reserved4"`
	Reserved5      string    `json:"-" gorm:"column:reserved5"`
	Updated        time.Time `json:"updated" gorm:"column:updated"`
	Created        time.Time `json:"created" gorm:"column:created"`
}

// Create machine problem table if it already does not exist (Database Migration)
func CreateMachineProblemTable() error {
	/*
		migration, err := qbs.GetMigration()
		if err != nil {
			stats.ERROR.Println("Failed to migrate machine problem database:", err)
			return err
		} else {
			stats.TRACE.Println("Created machine problem table")
		}
		defer migration.Close()
		return migration.CreateTableIfNotExists(new(MachineProblem))
	*/
	return nil
}

func FindMachineProblem(id int64) (MachineProblem, error) {
	var mp MachineProblem
	if err := DB.First(&mp, id).Error; err != nil {
		return mp, err
	}
	return mp, nil
}

func FindMachineProblemByUser(user User, mpNumber int) (MachineProblem, error) {
	var mp MachineProblem
	err := DB.
		Order("id DESC").
		First(&mp, MachineProblem{UserInstanceId: user.Id, Number: mpNumber}).
		Error
	if err != nil {
		return mp, err
	}
	return mp, nil
}

func FindMachineProblemsByNumber(mpNumber int) ([]MachineProblem, error) {
	var mps []MachineProblem

	err := DB.
		Order("id DESC").
		Find(&mps, MachineProblem{Number: mpNumber}).
		Error
	if err != nil {
		return mps, err
	}
	return mps, nil
}

func CreateMachineProblemForUser(user User, mpNumber int) (MachineProblem, error) {
	mp := MachineProblem{
		UserInstanceId: user.Id,
		Number:         mpNumber,
	}

	err := DB.Save(&mp).Error
	return mp, err
}

func FindOrCreateMachineProblemByUser(user User, mpNumber int) (MachineProblem, error) {
	if mp, err := FindMachineProblemByUser(user, mpNumber); err == nil {
		return mp, nil
	}
	return CreateMachineProblemForUser(user, mpNumber)
}

func RandomMachineProblemByNumber(mpNumber int) (MachineProblem, error) {
	var mp MachineProblem
	err := DB.
		Order("RAND()").
		First(&mp, MachineProblem{Number: mpNumber}).
		Error
	if err != nil {
		return mp, err
	}
	return mp, nil
}
