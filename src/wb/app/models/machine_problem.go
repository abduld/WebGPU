package models

import "time"

//
type MachineProblem struct {
	Id             int64     `json:"id" qbs:"pk,notnull"`
	Number         int       `json:"number" qbs:"notnull"`
	UserInstanceId int64     `json:"user_id" qbs:"fk:User"`
	Reserved1      string    `json:"-"`
	Reserved2      string    `json:"-"`
	Reserved3      string    `json:"-"`
	Reserved4      string    `json:"-"`
	Reserved5      string    `json:"-"`
	Updated        time.Time `json:"updated"`
	Created        time.Time `json:"created"`
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
		Where("user_instance_id = ? and number = ?", user.Id, mpNumber).
		Order("id DESC").
		First(&mp).
		Error
	if err != nil {
		return mp, err
	}
	return mp, nil
}

func FindMachineProblemsByNumber(mpNumber int) ([]MachineProblem, error) {
	var mps []MachineProblem
	err := DB.
		Where("number = ?", mpNumber).
		Order("id DESC").
		Find(&mps).
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
		Where("number = ?", mpNumber).
		Order("RAND()").
		First(&mp).
		Error
	if err != nil {
		return mp, err
	}
	return mp, nil
}
