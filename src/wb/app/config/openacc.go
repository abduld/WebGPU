package config

import (
	"github.com/robfig/revel"
)

var (
	PGCCCompilerLocation string
	PGCCCompiler string
)

func findPGCCDirectory() {
	if comp, found := NestedRevelConfig.String("openacc.compiler"); found {
		PGCCCompiler = comp
	} else {
		PGCCCompiler = "pgc++"
	}
	revel.TRACE.Println("PGCCCompiler = ", PGCCCompiler);
	if dir, found := NestedRevelConfig.String("openacc.directory"); found {
		PGCCCompilerLocation = dir
		revel.INFO.Println("Found OpenACC compiler ... ", PGCCCompilerLocation, "/", PGCCCompiler)
	} else {
		var err error
		PGCCCompilerLocation, err = findExe(PGCCCompiler)
		if err == nil {
			revel.INFO.Println("Found OpenACC compiler ... ", PGCCCompilerLocation, "/", PGCCCompiler)
		}
	}
}

func InitPGCCConfig() {
	findPGCCDirectory()
}
