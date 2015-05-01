package server

import (
	"path/filepath"

	. "wb/app/config"
)

func MakeCPPAMPCompileCommand(s *WorkerState) string {
	switch OperatingSystem {
	case "Linux":
		cmp := "#!/bin/sh\n" +
			CPPAMPCompiler +
			" `/opt/clamp/bin/clamp-config --install --cxxflags --ldflags`" +
			" -I " + filepath.Join(NVCCCompilerLocation, "..", "..", "include") +
			" " + filepath.Join(NVCCCompilerLocation, "..", "..", "lib64", "libcudart_static.a") +
			" -I" + CToolsDir +
			" -I" + filepath.Join(CUDAToolkitDirectory, "include") +
			" " + filepath.Join(CToolsDir, SystemId, "libwb.a") +
			" " + filepath.Join(s.TemporaryDirectory, s.ProgramFileName) +
			" -o " + filepath.Join(s.TemporaryDirectory, s.ExecutableFileName) +
			" -L." +
			" -lrt " +
			" -lcudart " +
			" -lstdc++ " +
			" -lm " +
			" -lcuda " +
			" -DWB_USE_CUSTOM_MALLOC " +
			" -DWB_USE_COURSERA " +
			" 2>&1\n"
		return cmp
	default:
		return "todo"
	}
}
