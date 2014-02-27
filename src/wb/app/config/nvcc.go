package config

import (
	. "wb/app/util"

	"github.com/robfig/revel"
)

var (
	NVCCCompilerLocation string
	CUDAToolkitDirectory string
)

func findNVCCDirectory() {
	if dir, found := NestedRevelConfig.String("nvcc.directory"); found {
		NVCCCompilerLocation = dir
	} else {
		NVCCCompilerLocation, _ = findExe("nvcc")
	}
}

func findCUDADirectory() {

	if dir, found := NestedRevelConfig.String("cuda.directory"); found {
		CUDAToolkitDirectory = dir
	} else if FileExists(NVCCCompilerLocation) {
		CUDAToolkitDirectory = ParentDirectory(ParentDirectory(NVCCCompilerLocation))
	} else {
		revel.ERROR.Println("Cannot find CUDAToolkitDirectory")
		if DebugMode {
			panic("Cannot find CUDAToolkitDirectory")
		}
	}
}

func InitNVCCConfig() {
	findNVCCDirectory()
	findCUDADirectory()
}
