package tests

import (
	"time"

	"github.com/abduld/oauth"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
	"github.com/revel/revel"
)

type UserTest struct {
	revel.TestSuite
	db gorm.DB
}

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
	RequestToken    *oauth.RequestToken `json:"-"`
	AccessToken     *oauth.AccessToken  `json:"-"`
	Reserved1       string              `json:"-"` // used for coursera's identity
	Reserved2       string              `json:"-"` // used for coursera's identity
	Reserved3       string              `json:"-"`
	Reserved4       string              `json:"-"`
	Reserved5       string              `json:"-"`
	Created         time.Time           `json:"-"`
	Updated         time.Time           `json:"-"`
}

const dbAddress = "wb:pass@tcp(localhost:8086)/wb?charset=utf8&parseTime=True"

func (t *UserTest) Before() {
	if db, err := gorm.Open("mysql", dbAddress); err == nil {
		db.LogMode(true)
		db.DB().SetMaxIdleConns(10)
		db.DB().SetMaxOpenConns(100)
		db.DB().Ping()

		db.SingularTable(true)
		t.db = db
	}
}

func (t *UserTest) After() {
}

func (t UserTest) TestUserQuery() {
	var user User
	db := t.db
	db.Find(&user)
	println("da")
}
