package config

import "github.com/robfig/revel"

var (
	DatabaseProvider   string
	DatabaseSourceName string
	DatabaseName       string
)

func InitDatabaseConfig() {

	conf := NestedRevelConfig

	provider, _ := conf.String("db.provider")
	database, _ := conf.String("db.database")
	host, _ := conf.String("db.host")
	port, _ := conf.String("db.port")
	user, _ := conf.String("db.user")
	password, _ := conf.String("db.password")

	DatabaseSourceName = user + ":" + password +
		"@tcp(" + host +
		":" + port +
		")/" + database +
		"?charset=utf8&parseTime=true&loc=Local"

	DatabaseName = database
	DatabaseProvider = provider

	revel.TRACE.Println(DatabaseProvider)

}
