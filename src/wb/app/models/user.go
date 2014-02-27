package models

import (
	"regexp"
	"time"

	"code.google.com/p/go.crypto/bcrypt"
	"github.com/abduld/oauth"
	"github.com/robfig/revel"
)

//
type User struct {
	Id              int64               `json:"id" qbs:"pk,notnull"`
	CourseraId      int64               `json:"coursera_id" qbs:"default:'-1'"` // not used, since coursera ids are strings
	FirstName       string              `json:"first_name" qbs:"default:''"`
	LastName        string              `json:"last_name" qbs:"default:''"`
	UserName        string              `json:"user_name" qbs:"default:''"`
	Email           string              `json:"email" qbs:"default:''"`
	Password        string              `json:"-" qbs:"default:''"`
	PasswordConfirm string              `json:"-" qbs:"default:''"`
	Affiliation     string              `json:"affiliation" qbs:"default:''"`
	TermsOfUse      bool                `json:"terms" qbs:"default:'0'"`
	Hashed          bool                `json:"-" qbs:"default:'0'"`
	RequestToken    *oauth.RequestToken `json:"-" sql:"-"`
	AccessToken     *oauth.AccessToken  `json:"-" sql:"-"`
	Reserved1       string              `json:"-"` // used for coursera's identity
	Reserved2       string              `json:"-"` // used for coursera's identity
	Reserved3       string              `json:"-"`
	Reserved4       string              `json:"-"`
	Reserved5       string              `json:"-"`
	Created         time.Time           `json:"-"`
	Updated         time.Time           `json:"-"`
}

// Create user table if it already does not exist (Database Migration)
func CreateUserTable() error {
	/*
		migration, err := qbs.GetMigration()
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
	return user, err
}

var findUserByNameCache map[string]User = map[string]User{}

func FindUserByName(name string) (User, error) {
	if user, ok := findUserByNameCache[name]; ok && user.Id != 0 {
		return user, nil
	}

	var user User

	err := DB.Where("user_name = ?", name).First(&user).Error
	if err == nil {
		findUserByNameCache[name] = user
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

func SetUserCourseraCredentials(user User, accessToken *oauth.AccessToken, requestToken *oauth.RequestToken,
	regularIdentity string, truestedIdentity string) (User, error) {

	user.RequestToken = requestToken
	user.AccessToken = accessToken
	user.Reserved1 = regularIdentity
	user.Reserved2 = truestedIdentity
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
