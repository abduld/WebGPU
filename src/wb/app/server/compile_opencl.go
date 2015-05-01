package server

import (
	"path/filepath"
	. "wb/app/config"
)

func MakeOpenCLCompileCommand(s *WorkerState) string {
	switch OperatingSystem {
	case "Linux":
		return "#!/bin/sh\n" +
			CCCompilerLocation +
			" -O3 " +
			" -x c++ " +
			" -Wl,-rpath=" + filepath.Join(CToolsDir, SystemId) +
			" " + filepath.Join(s.TemporaryDirectory, s.ProgramFileName) +
			" -o " + filepath.Join(s.TemporaryDirectory, s.ExecutableFileName) +
			" -DWB_USE_CUSTOM_MALLOC " +
			" -DWB_USE_COURSERA " +
			" -DWB_USE_CUDA " +
			" -DWB_USE_OPENCL " +
			" -lrt " +
			" -lstdc++ " +
			" -lm " +
			" -lcuda " +
			" -lOpenCL " +
			" -I" + CToolsDir +
			" -I" + filepath.Join(CUDAToolkitDirectory, "include") +
			//" " + filepath.Join(CToolsDir, SystemId, "libwb.a") +
			" -L " + filepath.Join(CToolsDir, SystemId) +
			" -lwb " +
			" 2>&1\n"
	default:
		return "todo"
	}
}
