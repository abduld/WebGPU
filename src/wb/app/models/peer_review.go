package models

import (
	"errors"
	"time"
	"wb/app/stats"
)

type PeerReview struct {
	Id                    int64     `json:"id" qbs:"pk,notnull"`
	GradeInstanceId       int64     `json:"grade_id" qbs:"fk:Grade"`
	Reviewer              int64     `json:"reviewer"`
	CodeReviewScore       int64     `json:"code_review_score"`
	CodeReviewComment     string    `json:"code_review_comment"`
	QuestionReviewScore   int64     `json:"question_review_score"`
	QuestionReviewComment string    `json:"question_review_comment"`
	Helpful               bool      `json:"helpful"`
	Reserved1             string    `json:"-"`
	Reserved2             string    `json:"-"`
	Reserved3             string    `json:"-"`
	Reserved4             string    `json:"-"`
	Reserved5             string    `json:"-"`
	Updated               time.Time `json:"updated"`
	Created               time.Time `json:"created"`
}

// Create peer review table if it already does not exist (Database Migration)
func CreatePeerReviewTable() error {
	/*
		migration, err := qbs.GetMigration()
		if err != nil {
			stats.ERROR.Println("Failed to create peer review database:", err)
			return err
		} else {
			stats.TRACE.Println("Created peer review table")
		}
		defer migration.Close()
		err = migration.CreateTableIfNotExists(new(PeerReview))
		return err
	*/
	return nil
}

func CreatePeerReview(grade Grade) (PeerReview, error) {
	pr := PeerReview{
		GradeInstanceId:       grade.Id,
		QuestionReviewComment: "",
		CodeReviewComment:     "",
		Created:               time.Now(),
		Updated:               time.Now(),
	}

	err := DB.Save(&pr).Error
	return pr, err
}

func FindPeerReview(id int64) (PeerReview, error) {
	var pr PeerReview

	if err := DB.Find(&pr, id).Error; err != nil {
		return pr, err
	}
	return pr, nil
}

func FindPeerReviewsWithGrade(grade Grade) ([]PeerReview, error) {
	var prs []PeerReview
	err := DB.
		Where("grade_instance_id = ?", grade.Id).
		Order("id DESC").
		Find(&prs).
		Error
	if err != nil {
		return prs, err
	} else if len(prs) == 0 {
		return prs, errors.New("cannot find peer review by grade.")
	}

	var set map[int64]PeerReview = map[int64]PeerReview{}
	for _, pr := range prs {
		if _, ok := set[pr.GradeInstanceId]; !ok {
			set[pr.GradeInstanceId] = pr
		}
	}

	prs = []PeerReview{}
	for _, val := range set {
		prs = append(prs, val)
	}

	return prs, err
}

func CreatePeerReviewWithReviewerAndMachineProblem(user User, mp MachineProblem) (PeerReview, error) {
	var pr PeerReview

	for ii := 0; ii < 1000; ii++ {
		useQ := true
		grade, err := RandomGradeByMachineProblem(mp.Number)
		if err != nil {
			continue
		}
		tmp, err := FindMachineProblem(grade.MachineProblemId)
		if err != nil || tmp.UserInstanceId == user.Id {
			continue
		}
		prs, err := FindPeerReviewsWithGrade(grade)
		if err == nil && len(prs) >= 3 {
			continue
		}
		if err == nil {
			for _, pr := range prs {
				if pr.Reviewer == user.Id {
					useQ = false
				}
			}
		}
		if !useQ {
			continue
		}
		pr := PeerReview{
			GradeInstanceId:       grade.Id,
			QuestionReviewComment: "",
			CodeReviewComment:     "",
			Reviewer:              user.Id,
			Created:               time.Now(),
			Updated:               time.Now(),
		}
		err = DB.
			Save(&pr).
			Error
		return pr, err
	}
	return pr, errors.New("Cannot create peer review.")
}

func GetPeerReviewAttempt(pr PeerReview) (Attempt, error) {
	if grade, err := FindGrade(pr.GradeInstanceId); err == nil {
		if attempt, err := FindAttempt(grade.AttemptInstanceId); err == nil {
			return attempt, nil
		}
	}
	return Attempt{}, errors.New("Cannot find attempt")
}

func GetPeerReviewReviewerUser(pr PeerReview) (User, error) {
	return FindUser(pr.Reviewer)
}

func GetPeerReviewUser(pr PeerReview) (User, error) {
	mp, err := GetPeerReviewMachineProblem(pr)
	if err == nil {
		return FindUser(mp.UserInstanceId)
	}
	return User{}, err
}

func GetPeerReviewMachineProblem(pr PeerReview) (MachineProblem, error) {
	if grade, err := FindGrade(pr.GradeInstanceId); err == nil {
		return FindMachineProblem(grade.MachineProblemId)
	}
	return MachineProblem{}, nil
}

func GetPeerReviewsByReviewerAndMachineProblem(user User, mp MachineProblem) ([]PeerReview, error) {
	var res []PeerReview
	var prs []PeerReview

	err := DB.
		Where("reviewer = ?", user.Id).
		Order("id DESC").
		Find(&prs).
		Error

	if err != nil {
		stats.TRACE.Println("Cannot find peer reviews.")
		return nil, err
	}

	var set map[int64]bool = map[int64]bool{}

	for _, p := range prs {
		mpPr, err := GetPeerReviewMachineProblem(p)
		if err == nil && mp.Number == mpPr.Number {
			if _, ok := set[p.GradeInstanceId]; !ok {
				res = append([]PeerReview{p}, res...)
				set[p.GradeInstanceId] = true
			}
		}
	}

	if len(res) == 0 {
		stats.TRACE.Println("Cannot find peer reviews.")
		return nil, errors.New("Cannot find any peer reviews associated with reviewer.")
	}
	return res, nil
}

func GetQuestionsByPeerReview(pr PeerReview) (Questions, error) {
	mp, err := GetPeerReviewMachineProblem(pr)
	if err != nil {
		return Questions{}, err
	}
	return FindQuestionsByMachineProblem(mp)
}

func NumberOfPeerReviewsByUserAndMachineProblem(user User, mp MachineProblem) int {
	if prs, err := GetPeerReviewsByReviewerAndMachineProblem(user, mp); err == nil {
		return len(prs)
	}
	return 0
}

func GetPeerReviewsByReviewer(user User) ([]PeerReview, error) {
	var pr []PeerReview
	err := DB.
		Where("reviewer = ?", user.Id).
		Order("id DESC").
		Find(&pr).
		Error
	if err != nil {
		return pr, err
	} else if len(pr) == 0 {
		return pr, errors.New("Cannot find peer reviews.")
	}

	return pr, err
}

func UpdatePeerReview(pr PeerReview, codeReviewScore int64, codeReviewComment string,
	questionReviewScore int64, questionReviewComment string) (PeerReview, error) {

	pr.Id = 0
	pr.CodeReviewScore = codeReviewScore
	pr.CodeReviewComment = codeReviewComment
	pr.QuestionReviewScore = questionReviewScore
	pr.QuestionReviewComment = questionReviewComment
	pr.Updated = time.Now()

	err := DB.Save(&pr).Error
	return pr, err
}
