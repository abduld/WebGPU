package models

import (
	"errors"
	"sync"
	"time"
	. "wb/app/config"
	"wb/app/stats"
)

type QuestionItem struct {
	Id                  int64     `json:"id" qbs:"pk,notnull"`
	Number              int64     `json:"number"`
	QuestionsInstanceId int64     `json:"question_id" qbs:"fk:Questions"`
	QuestionText        string    `json:"question"`
	Answer              string    `json:"answer qbs:"default:''"`
	Reserved1           string    `json:"-"`
	Reserved2           string    `json:"-"`
	Reserved3           string    `json:"-"`
	Reserved4           string    `json:"-"`
	Reserved5           string    `json:"-"`
	Updated             time.Time `json:"updated"`
	Created             time.Time `json:"created"`
}

var questionsMutexs map[int64]*sync.Mutex = map[int64]*sync.Mutex{}

type Questions struct {
	Id                       int64     `json:"id" qbs:"pk,notnull"`
	MachineProblemInstanceId int64     `json:"-" qbs:"fk:machine_problem"`
	Reserved1                string    `json:"-"`
	Reserved2                string    `json:"-"`
	Reserved3                string    `json:"-"`
	Reserved4                string    `json:"-"`
	Reserved5                string    `json:"-"`
	Updated                  time.Time `json:"updated"`
	Created                  time.Time `json:"created"`
}

func CreateQuestionsTable() error {
	/*
		migration, err := qbs.GetMigration()
		if err != nil {
			return err
		}
		defer migration.Close()
		err = migration.CreateTableIfNotExists(new(Questions))
		if err != nil {
			return err
		}
		return migration.CreateTableIfNotExists(new(QuestionItem))
	*/
	return nil
}

func CreateQuestion(mp MachineProblem, q Questions, conf *MachineProblemConfig, num int64) (QuestionItem, error) {
	r := conf.Questions[num]
	questionItem := QuestionItem{
		Number:              num + 1,
		QuestionsInstanceId: q.Id,
		QuestionText:        r,
		Answer:              "",
		Created:             time.Now(),
		Updated:             time.Now(),
	}
	err := DB.Save(&questionItem).Error
	return questionItem, err
}

func CreateQuestions(mp MachineProblem) (Questions, error) {
	conf, err := ReadMachineProblemConfig(mp.Number)

	if err != nil {
		return Questions{}, err
	}

	q := Questions{
		MachineProblemInstanceId: mp.Id,
		Created:                  time.Now(),
		Updated:                  time.Now(),
	}
	err = DB.Save(&q).Error
	if err != nil {
		return q, err
	}

	for i := range conf.Questions {
		CreateQuestion(mp, q, conf, int64(i))
	}

	return q, nil
}

// may include duplicates
func FindAllQuestionItems(q Questions) ([]QuestionItem, error) {
	var qis []QuestionItem
	err := DB.
		Where("questions_instance_id = ?", q.Id).
		Order("updated DESC").
		Order("number").
		Find(&qis).
		Error
	if err != nil {
		return qis, err
	} else if len(qis) == 0 {
		return qis, errors.New("No questions")
	}
	return qis, nil
}

func FindQuestionHistory(q Questions, n int64) ([]QuestionItem, error) {

	var qis []QuestionItem
	err := DB.
		Where("questions_instance_id = ? and number = ?", q.Id, n).
		Order("updated DESC").
		Find(&qis).
		Error
	if err != nil {
		return qis, err
	} else if len(qis) == 0 {
		return qis, errors.New("No questions")
	}

	return qis, nil
}

func FindQuestionItems(mp MachineProblem, q Questions) ([]QuestionItem, error) {
	qs, err := FindAllQuestionItems(q)
	if err != nil {
		return nil, err
	}

	conf, err := ReadMachineProblemConfig(mp.Number)

	mpQuestions := conf.Questions

	res := make([]QuestionItem, len(mpQuestions))
	for _, q := range qs {
		if q.Id != 0 && (q.Number) <= int64(len(res)) && res[q.Number-1].Id == 0 {
			res[q.Number-1] = q
		}
	}
	for ii, r := range res {
		if r.Id == 0 {
			if q, err := CreateQuestion(mp, q, conf, int64(ii)); err == nil {
				res[ii] = q
			}
		}
	}

	return res, err
}

func findQuestionItem(mp MachineProblem, q Questions, num int64) (QuestionItem, error) {
	qs, err := FindQuestionItems(mp, q)
	if err != nil {
		return QuestionItem{}, errors.New("Cannot find questions")
	}
	for _, p := range qs {
		if p.Id != 0 && p.Number == num {
			p.QuestionsInstanceId = q.Id
			return p, err
		}
	}
	return QuestionItem{}, errors.New("Cannot find question")
}

func SaveQuestion(mp MachineProblem, qs Questions, num int64, answer string) error {
	q, err := findQuestionItem(mp, qs, num)
	if err != nil {
		return err
	}

	conf, err := ReadMachineProblemConfig(mp.Number)
	mpQuestions := conf.Questions

	newQuestion := QuestionItem{
		Number:              q.Number,
		QuestionsInstanceId: qs.Id,
		QuestionText:        mpQuestions[q.Number-1],
		Answer:              answer,
		Created:             qs.Created,
		Updated:             time.Now(),
	}

	err = DB.Save(&newQuestion).Error

	return err
}

func SaveQuestions(mp MachineProblem, qs Questions, answers []string) error {
	for r := range answers {
		err := SaveQuestion(mp, qs, int64(r), answers[r])
		if err != nil {
			stats.INFO.Println("Faield to save question qid = ", qs.Id)
		}
	}
	return nil
}

func FindQuestions(id int64) (Questions, error) {
	var questions Questions
	err := DB.
		Where("id = ?", id).
		Order("updated DESC").
		First(questions).
		Error
	if err != nil {
		return questions, err
	}
	return questions, err
}

func FindQuestionsByMachineProblem(mp MachineProblem) (Questions, error) {
	var questions Questions
	err := DB.
		Where("machine_problem_instance_id = ?", mp.Id).
		Order("updated DESC").
		First(&questions).
		Error
	if err != nil {
		return questions, err
	} else if questions.Id == 0 {
		return questions, errors.New("Question not found")
	}
	questions.MachineProblemInstanceId = mp.Id
	return questions, err
}

func FindOrCreateQuestionsByMachineProblem(mp MachineProblem) (Questions, error) {
	questions, err := FindQuestionsByMachineProblem(mp)
	if err != nil {
		return CreateQuestions(mp)
	}
	return questions, err
}
