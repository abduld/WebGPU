package config

import (
	"runtime"

	"github.com/revel/revel"
)

var (
	BigCodeEnabled         bool
	BigCodeVersion         int
	BigCodeClangPath       string
	BigCodeOptPath         string
	BigCodeLibraryPath     string
	BigCodeNumberOfBatches int
)

func findBigCodeClangPath() {
	if dir, found := NestedRevelConfig.String("bigcode.clang_path"); found {
		BigCodeClangPath = dir
	} else {
		revel.ERROR.Println("Cannot find BigCode Clang Path")
		if DebugMode {
			panic("Cannot find BigCode Clang Path")
		}
	}
}

func findBigCodeOptPath() {
	if dir, found := NestedRevelConfig.String("bigcode.llvm_opt_path"); found {
		BigCodeOptPath = dir
	} else {
		revel.ERROR.Println("Cannot find BigCode LLVM Opt Path")
		if DebugMode {
			panic("Cannot find BigCode LLVM Opt Path")
		}
	}
}

func findBigCodeLibraryPath() {
	if dir, found := NestedRevelConfig.String("bigcode.library_path"); found {
		BigCodeLibraryPath = dir
	} else {
		revel.ERROR.Println("Cannot find BigCode Library Path")
		if DebugMode {
			panic("Cannot find BigCode Library Path")
		}
	}
}

func InitBigCodeConfig() {
	if val, found := NestedRevelConfig.Bool("bigcode.enabled"); found {
		BigCodeEnabled = val
	} else {
		BigCodeEnabled = IsWorker
	}
	if n, found := NestedRevelConfig.Int("bigcode.version"); found {
		BigCodeVersion = n
	} else {
		BigCodeVersion = 0
	}
	if n, found := NestedRevelConfig.Int("bigcode.number_of_batches"); found {
		BigCodeNumberOfBatches = n
	} else {
		BigCodeNumberOfBatches = 2*runtime.NumCPU() - 1
	}
	findBigCodeClangPath()
	findBigCodeOptPath()
	findBigCodeLibraryPath()
}
