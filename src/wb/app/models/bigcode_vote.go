package models

import (
	"time"
)

type BigcodeVote struct {
	Id                        int64     `json:"id" gorm:"column:id; primary_key:yes;"`
	UserProgramId             int64     `json:"user_program_id" gorm:"column:user_program_id;"`
	UserProgramSignature      string    `json:"-" gorm:"column:user_program_signature;"`
	SuggestedProgramId        int64     `json:"user_program_id" gorm:"column:suggested_program_id;"`
	SuggestedProgramSignature string    `json:"-" gorm:"column:suggested_program_signature;"`
	DistanceMetric            string    `json:"distance_metric" gorm:"column:distance_metric;"`
	SuggestionType            string    `json:"suggestion_type" gorm:"column:suggestion_type;"`
	Vote                      bool      `json:"vote" gorm:"column:vote;"`
	Created                   time.Time `json:"created" gorm:"column:created"`
	Reserved1                 string    `json:"-" gorm:"column:reserved1"`
	Reserved2                 string    `json:"-" gorm:"column:reserved2"`
	Reserved3                 string    `json:"-" gorm:"column:reserved3"`
	Reserved4                 string    `json:"-" gorm:"column:reserved4"`
	Reserved5                 string    `json:"-" gorm:"column:reserved5"`
}

func CreateBigcodeVoteTable() error {
	return nil
}
func CreateBigcodeVote(vote BigcodeVote) (BigcodeVote, error) {
	v := vote
	err := DB.Save(&v).Error
	return v, err
}
