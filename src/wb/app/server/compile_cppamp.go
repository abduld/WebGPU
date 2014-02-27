package server

import (
	"path/filepath"
	. "wb/app/config"
)

func MakeCPPAMPCompileCommand(s *WorkerState) string {
	switch OperatingSystem {
	case "Linux":
		return "#!/bin/sh\n" +
			CPPAMPCompilerLocation +
			" -O3 " +
			" -lrt " +
			" -lstdc++ " +
			" -lm " +
			" -lcppamp " +
			" -I" + CToolsDir +
			" " + filepath.Join(CToolsDir, SystemId, "libwb.a") +
			" " + filepath.Join(s.TemporaryDirectory, s.ProgramFileName) +
			" -o " + filepath.Join(s.TemporaryDirectory, s.ExecutableFileName) +
			" -DWB_USE_CUSTOM_MALLOC " +
			" -DWB_USE_COURSERA " +
			" 2>&1\n"
	default:
		return "todo"
	}
}
