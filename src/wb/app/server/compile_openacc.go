package server

import (
	"path/filepath"
	. "wb/app/config"
)

func MakeOpenACCCompileCommand(s *WorkerState) string {
	switch OperatingSystem {
	case "Linux":
		return "#!/bin/sh\n" +
			PGCCCompilerLocation +
			" -O3 " +
			" -acc " +
			" -ta=nvidia " +
			" -Minfo=accel " +
			" -lrt " +
			" -lstdc++ " +
			" -lm " +
			" -lcuda " +
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
