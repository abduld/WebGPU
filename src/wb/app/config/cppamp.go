package config

import (
	"github.com/robfig/revel"
)

var (
	CPPAMPCompilerLocation string
)

func findCPPAMPDirectory() {
	revel.WARN.Println("Cannot find CPPAMPCompilerLocation")
	// panic("Cannot find CPPAMPCompilerLocation")
}

func InitCPPAMPConfig() {
	findCPPAMPDirectory()
}
