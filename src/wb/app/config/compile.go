package config

import (
	"path/filepath"
)

var (
	JanssonDir string
	CToolsDir  string
)

func InitCompileConfig() {
	CToolsDir = filepath.Join(BasePath, "c-tools")
	JanssonDir = filepath.Join(CToolsDir, "jansson")
}
