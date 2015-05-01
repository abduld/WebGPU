package models

import (
	"strings"
	"time"
	. "wb/app/config"

	"github.com/revel/revel"
)

var (
	ECE408AdminUsers   []User
	ECE598AdminUsers   []User
	ECE408StudentUsers []User
	ECE598StudentUsers []User
	AllECEStudentUsers []User
)

func inEmailList(user User, emails []string, kind string) bool {
	userEmail := strings.TrimSpace(strings.ToLower(user.Email))
	for _, email := range emails {
		if email == userEmail {
			return true
		}
	}
	return false
}

func inUserList(user User, lst []User) bool {
	for _, elem := range lst {
		if elem.Id == user.Id {
			return true
		}
	}
	return false
}

func IsECE408Admin(user User) bool {
	return user.UserName == "admin" || inEmailList(user, ECE408Admins, "ECE408Admins")
}

func IsECE598Admin(user User) bool {
	return user.UserName == "admin" || inEmailList(user, ECE598Admins, "ECE598Admins")
}

func IsECE408Student(user User) bool {
	return inEmailList(user, ECE408Students, "ECE408Students") || IsECE408Admin(user)
}

func IsECE598Student(user User) bool {
	if strings.Contains(user.Email, "@ou.edu") || strings.Contains(user.Email, "@utk.edu") {
		return true
	}
	return inEmailList(user, ECE598Students, "ECE598Students") || IsECE598Admin(user)
}

func IsAdmin(user User) bool {
	return IsECE408Admin(user) || IsECE598Admin(user)
}

func unionUserList(users1 []User, users2 []User) []User {
	tmp := map[int64]User{}
	for _, val := range users1 {
		tmp[val.Id] = val
	}
	for _, val := range users2 {
		tmp[val.Id] = val
	}
	res := make([]User, 0, len(tmp))
	for _, v := range tmp {
		res = append(res, v)
	}
	return res
}

func getUsersByEmail(className string, userEmailList []string) ([]User, error) {
	count := 0
	for _, email := range userEmailList {
		if _, err := FindUserByEmail(email); err == nil {
			count++
		}
	}
	users := make([]User, count)
	for _, email := range userEmailList {
		if user, err := FindUserByEmail(email); err == nil {
			users = append(users, user)
			revel.TRACE.Println("Found " + className + " record for " + email)
		} else {
			revel.TRACE.Println("Cannot find " + className + " record for " + email)
		}
	}
	return users, nil
}
func readUsers() {
	if val, err := getUsersByEmail("ECE408", ECE408Students); err == nil {
		ECE408StudentUsers = val
	}
	if val, err := getUsersByEmail("ECE598", ECE598Students); err == nil {
		ECE598StudentUsers = val
	}
	if val, err := getUsersByEmail("ECE408", ECE408Admins); err == nil {
		ECE408AdminUsers = val
	}
	if val, err := getUsersByEmail("ECE598", ECE598Admins); err == nil {
		ECE598AdminUsers = val
	}
	AllECEStudentUsers = unionUserList(ECE408StudentUsers, ECE598StudentUsers)
}

func getMPForUser(mpNum int, email string) (MachineProblem, error) {
	if len(AllECEStudentUsers) == 0 {
		readUsers()
	}
	user, err := FindUserByEmail(email)
	if err != nil {
		return MachineProblem{}, err
	}
	return FindMachineProblemByUser(user, mpNum)
}

func setUserGrade(mpNum int, email string, codeGrade int64, codeComments string,
	questionsGrade int64, questionComment string) (Grade, error) {
	mp, err := getMPForUser(mpNum, email)
	if err != nil {
		return Grade{}, err
	}
	grade, err := FindGradeByMachineProblem(mp)
	if err != nil {
		return Grade{}, err
	}

	grade.CodeScore = codeGrade
	grade.PeerReviewScore = questionsGrade
	grade.Reserved4 = codeComments
	grade.Reserved5 = questionComment
	grade.TotalScore = codeGrade + questionsGrade

	grade.Updated = time.Now()

	err = DB.Save(&grade).Error
	return grade, err
}
