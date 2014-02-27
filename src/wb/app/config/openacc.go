package config

import (
	"github.com/robfig/revel"
)

var (
	PGCCCompilerLocation string
)

func findPGCCDirectory() {

	if dir, found := NestedRevelConfig.String("pgcc.directory"); found {
		PGCCCompilerLocation = dir
	} else if exe, err := findExe("pgcc"); err == nil {
		PGCCCompilerLocation = exe
	} else {
		revel.WARN.Println("Cannot find PGCCCompilerLocation")
	}
}

func InitPGCCConfig() {
	findPGCCDirectory()
}
