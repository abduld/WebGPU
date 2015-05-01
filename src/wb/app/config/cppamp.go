package config

import "path/filepath"

var (
	CPPAMPCompilerLocation string
	CPPAMPCompiler         string
	CPPAMPConfig           string
)

func findCPPAMPDirectory() {
	if dir, found := NestedRevelConfig.String("cppamp.directory"); found {
		CPPAMPCompilerLocation = dir
	} else {
		CPPAMPCompilerLocation = "/opt/clamp"
	}
	if comp, found := NestedRevelConfig.String("cppamp.compiler"); found {
		CPPAMPCompiler = comp
	} else {
		CPPAMPCompiler = filepath.Join(CPPAMPCompilerLocation, "bin", "clang++")
	}
	if conf, found := NestedRevelConfig.String("cppamp.config"); found {
		CPPAMPConfig = conf
	} else {
		CPPAMPConfig = filepath.Join(CPPAMPCompilerLocation, "bin", "clamp-config")
	}
}

func InitCPPAMPConfig() {
	findCPPAMPDirectory()
}
