package models

import (
	"errors"
	"regexp"
	"sort"
	"strings"
	"time"

	"code.google.com/p/go.crypto/bcrypt"
	"golang.org/x/oauth2"

	"github.com/revel/revel"
)

//
type User struct {
	Id              int64         `json:"id" gorm:"column:id; primary_key:yes"`
	CourseraId      int64         `json:"-" gorm:"column:coursera_id" sql:"default:'-1'"` // not used, since coursera ids are strings
	FirstName       string        `json:"first_name" gorm:"column:first_name" sql:"default:''"`
	LastName        string        `json:"last_name" gorm:"column:last_name" sql:"default:''"`
	UserName        string        `json:"user_name" gorm:"column:user_name" sql:"default:''"`
	Email           string        `json:"email" gorm:"column:email" sql:"default:''"`
	Password        string        `json:"-" gorm:"column:password" sql:"default:''"`
	PasswordConfirm string        `json:"-" gorm:"column:password_confirm" sql:"default:''"`
	Affiliation     string        `json:"affiliation" gorm:"column:affiliation" sql:"default:''"`
	TermsOfUse      bool          `json:"terms" gorm:"column:terms_of_use" sql:"default:'0'"`
	Hashed          bool          `json:"-" gorm:"column:hashed" sql:"default:'0'"`
	RequestToken    *oauth2.Token `json:"-" sql:"-"`
	Reserved1       string        `json:"r1" gorm:"column:reserved1"` // used for coursera's identity
	Reserved2       string        `json:"r2" gorm:"column:reserved2"` // used for coursera's identity
	Reserved3       string        `json:"-" gorm:"column:reserved3"`
	Reserved4       string        `json:"-" gorm:"column:reserved4"`
	Reserved5       string        `json:"-" gorm:"column:reserved5"`
	Updated         time.Time     `json:"updated" gorm:"column:updated"`
	Created         time.Time     `json:"created" gorm:"column:created"`
}

// Create user table if it already does not exist (Database Migration)
func CreateUserTable() error {
	/*
		migration, err := sql.GetMigration()
		if err != nil {
			stats.ERROR.Println("Failed to migrate user database:", err)
			return err
		} else {
			stats.TRACE.Println("Created user table")
		}
		defer migration.Close()
		return migration.CreateTableIfNotExists(new(User))
	*/
	return nil
}

func FindUser(id int64) (User, error) {
	var user User
	err := DB.First(&user, id).Error
	if user.Id == 0 {
		return User{}, errors.New("Invalid user")
	}
	return user, err
}

var findUserByNameCache map[string]User = map[string]User{}
var findUserByEmailCache map[string]User = map[string]User{}

func FindUserByName(name string) (User, error) {
	if user, ok := findUserByNameCache[name]; ok && user.Id != 0 {
		return user, nil
	}
	var user User

	err := DB.
		First(&user, User{UserName: name}).
		Error
	if err == nil {
		findUserByNameCache[name] = user
	}

	return user, err
}

func FindUserByMachineProblem(mpId int64) (User, error) {
	mp, err := FindMachineProblem(mpId)
	if err != nil {
		return User{}, err
	}
	return FindUser(mp.UserInstanceId)
}

func stringLess(s, t string) int {
	a := []byte(s)
	b := []byte(t)
	for i := 0; i < len(a) && i < len(b); i++ {
		switch {
		case a[i] > b[i]:
			return 1
		case a[i] < b[i]:
			return -1
		}
	}
	switch {
	case len(a) > len(b):
		return 1
	case len(a) < len(b):
		return -1
	}
	return 0
}

func stringLessQ(s, t string) bool {
	return stringLess(s, t) <= 0
}

type ByEmail []User

func (a ByEmail) Len() int      { return len(a) }
func (a ByEmail) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a ByEmail) Less(i, j int) bool {
	return stringLessQ(
		strings.ToLower(a[i].Email),
		strings.ToLower(a[j].Email))
}

func inplaceSortUsersByEmail(users []User) {
	sort.Sort(ByEmail(users))
}

func FindUsersByMachineProblemNumber(mpNumber int) ([]User, error) {
	users := []User{}
	err := DB.
		Joins("inner join machine_problem mp on mp.user_instance_id = user.id").
		Where("mp.number = ?", mpNumber).
		Order("user.email").
		Select("user.*").
		Find(&users).
		Error

	return users, err
}

func FindUserByEmail(email string) (User, error) {
	if user, ok := findUserByEmailCache[email]; ok && user.Id != 0 {
		return user, nil
	}

	var user User
	var users []User

	err := DB.
		Order("id DESC").
		Find(&users, User{Email: email}).
		Error
	if err == nil {
		if len(users) == 0 {
			err = errors.New("No user found")
		} else {
			user = users[0]
			findUserByEmailCache[email] = user
		}
	}

	return user, err
}

func UserExists(id int64) bool {
	_, err := FindUser(id)
	return err == nil
}

func UserNameExists(name string) bool {
	_, err := FindUserByName(name)
	return err == nil
}

func ValidUserNamePassword(name string, pass string) bool {
	user, err := FindUserByName(name)
	if err == nil {
		if samePassword(user.Password, pass) {
			return true
		}
	}
	return false
}

func CreateUser(user User) error {
	if user.Hashed == false {
		user.Password = hashPassword(user.Password)
		user.PasswordConfirm = hashPassword(user.PasswordConfirm)
		user.Hashed = true
	}
	user.Created = time.Now()
	user.Updated = time.Now()
	err := DB.Save(&user).Error
	return err
}

var userRegex = regexp.MustCompile("^\\w*$")

func (user *User) Validate(v *revel.Validation) {
	v.Check(user.UserName,
		revel.Required{},
		revel.MaxSize{36},
		revel.MinSize{4},
		revel.Match{userRegex},
	).Message("User name must be between 4 and 36 alphanumeric characters")
	v.Required(user.FirstName).
		Message("First name is required")
	v.Required(user.LastName).
		Message("Last name is required")
	v.Required(user.Password).
		Message("Password is required")
	v.Required(user.PasswordConfirm == user.Password).
		Message("The passwords do not match.")
	validatePassword(v, user.Password).Key("user.Password").
		Message("Password must be between 5 and 65 characters long")

	v.Required(user.Email).
		Message("Email is required")
	v.Email(user.Email).
		Message("Email must be valid")
	v.Required(user.TermsOfUse).
		Message("Must agree to honor code.")
}

func validatePassword(v *revel.Validation, password string) *revel.ValidationResult {
	return v.Check(password,
		revel.Required{},
		revel.MaxSize{65},
		revel.MinSize{5},
	)
}

func SetUserCourseraIdentity(user User, requestToken *oauth2.Token,
	publicid string) (User, error) {

	user.RequestToken = requestToken
	user.Reserved3 = publicid
	user.Updated = time.Now()
	err := DB.Save(&user).Error
	delete(findUserByNameCache, user.UserName)
	return user, err
}

func SetUserCourseraCredentials(user User,
	value string, form string, userid string) (User, error) {

	user.Reserved2 = value + form
	user.Reserved1 = userid
	user.Updated = time.Now()
	err := DB.Save(&user).Error
	delete(findUserByNameCache, user.UserName)
	return user, err
}

func GetUserTrustedIdentity(user User) string {
	return user.Reserved2
}

func GetUserIdentity(user User) string {
	return user.Reserved1
}

func hashPassword(pass string) string {
	hash, _ := bcrypt.GenerateFromPassword([]byte(pass), bcrypt.DefaultCost)
	return string(hash)
}

func samePassword(hash string, pass string) bool {
	hashb := []byte(hash)
	passb := []byte(pass)
	return bcrypt.CompareHashAndPassword(hashb, passb) == nil
}

func ResetUserPassword(user User, pass string) error {
	user.Password = hashPassword(pass)
	user.PasswordConfirm = hashPassword(pass)
	user.Updated = time.Now()
	err := DB.Save(&user).Error
	delete(findUserByNameCache, user.UserName)
	return err
}
