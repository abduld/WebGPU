package config

import (
	"os/exec"

	"github.com/revel/revel"
)

var (
	CCCompilerLocation string
)

func findExe(s string) (string, error) {
	ccName := s
	if OperatingSystem == "Windows" {
		ccName += ".exe"
	}
	if res, err := exec.LookPath(ccName); err == nil {
		return res, err
	} else {
		revel.WARN.Println("Cannot find CompilerLocation", ccName)
		return "", err
	}
}

func findCCCompilerDirectory() {
	if dir, found := NestedRevelConfig.String("ccompiler.directory"); found {
		CCCompilerLocation = dir
	} else {
		var err error
		CCCompilerLocation, err = findExe("gcc")
		if err == nil {
			revel.INFO.Println("Found gcc compiler ... ", CCCompilerLocation)
		}
	}
}

func InitCCConfig() {
	findCCCompilerDirectory()
}
