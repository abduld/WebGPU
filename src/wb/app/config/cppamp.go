package config

import (
	"github.com/robfig/revel"
)

var (
	CPPAMPCompilerLocation string
	CPPAMPCompiler string
)

func findCPPAMPDirectory() {
	if comp, found := NestedRevelConfig.String("openacc.compiler"); found {
		CPPAMPCompiler = comp
	} else {
		CPPAMPCompiler = "pgcpp"
	}
	if dir, found := NestedRevelConfig.String("openacc.directory"); found {
		CPPAMPCompilerLocation = dir
	} else {
		var err error
		CPPAMPCompilerLocation, err = findExe(CPPAMPCompiler)
		if err == nil {
			revel.INFO.Println("Found OpenACC compiler ... ", CPPAMPCompilerLocation)
		}
	}
}

func InitCPPAMPConfig() {
	findCPPAMPDirectory()
}
