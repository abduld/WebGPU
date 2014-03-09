package server

import (
	"path/filepath"
	. "wb/app/config"
)

func MakeOpenCLCompileCommand(s *WorkerState) string {
	switch OperatingSystem {
	case "Linux":
		return "#!/bin/sh\n" +
			NVCCCompilerLocation +
			" -O3 " +
			" -lrt " +
			" -lstdc++ " +
			" -lm " +
			" -lcuda " +
			" -lOpenCL " +
			" -I" + CToolsDir +
			" -I" + filepath.Join(CUDAToolkitDirectory, "include") +
			" " + filepath.Join(CToolsDir, SystemId, "libwb.a") +
			" " + filepath.Join(s.TemporaryDirectory, s.ProgramFileName) +
			" -o " + filepath.Join(s.TemporaryDirectory, s.ExecutableFileName) +
			" -DWB_USE_CUSTOM_MALLOC " +
			" -DWB_USE_COURSERA " +
			" -DWB_USE_CUDA " +
			" -DWB_USE_OPENCL " +
			" 2>&1\n"
	default:
		return "todo"
	}
}
