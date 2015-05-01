package server

import (
	"path/filepath"
	. "wb/app/config"
)

func MakeNVCCCompileCommand(s *WorkerState) string {
	switch OperatingSystem {
	case "Linux":
		return "#!/bin/sh\n" +
			NVCCCompilerLocation +
			" -O3 " +
			" -lrt " +
			" -lstdc++ " +
			" -lm " +
			" -lcuda " +
			" -lcudadevrt " +
			" -lcudart " +
			" -arch=sm_30 " +
			" -I" + CToolsDir +
			//" " + filepath.Join(CToolsDir, SystemId, "libwb.a") +
			" -L " + filepath.Join(CToolsDir, SystemId) +
			" -lwb " +
			" -Xlinker -rpath=" + filepath.Join(CToolsDir, SystemId) +
			" " + filepath.Join(s.TemporaryDirectory, s.ProgramFileName) +
			" -o " + filepath.Join(s.TemporaryDirectory, s.ExecutableFileName) +
			" -DWB_USE_CUSTOM_MALLOC " +
			" -DWB_USE_COURSERA " +
			" -DWB_USE_CUDA " +
			" 2>&1\n"
	case "MacOSX":
		return "#!/bin/sh\n" +
			NVCCCompilerLocation +
			" -O3 " +
			" -lm " +
			" -lcuda " +
			" -lcudart " +
			" -arch=sm_30 " +
			" -L/usr/local/cuda/lib" +
			" -I/usr/local/cuda/include" +
			" -I" + CToolsDir +
			" " + filepath.Join(CToolsDir, SystemId, "libwb.a") +
			//" -L " + filepath.Join(CToolsDir, SystemId) +
			//" -lwb " +
			//" -Xlinker -rpath=" + filepath.Join(CToolsDir, SystemId) +
			" " + filepath.Join(s.TemporaryDirectory, s.ProgramFileName) +
			" -o " + filepath.Join(s.TemporaryDirectory, s.ExecutableFileName) +
			" -DWB_USE_CUSTOM_MALLOC " +
			" -DWB_USE_COURSERA " +
			" -DWB_USE_CUDA " +
			" 2>&1\n"
	default:
		return "todo"
	}
}
