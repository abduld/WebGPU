package config

import (
	"path/filepath"
)

var (
	GeoIPDatabaseFile string
)

func InitGeoIPConfig() {
	geodb, _ := NestedRevelConfig.String("geoip.db")
	GeoIPDatabaseFile = filepath.Join(BasePath, geodb)
}
