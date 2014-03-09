package server

import (
	"path/filepath"
	
	"github.com/robfig/revel"
	. "wb/app/config"
)

func MakeCPPAMPCompileCommand(s *WorkerState) string {
	switch OperatingSystem {
	case "Linux":
		cmp := "#!/bin/sh\n" +
			CPPAMPCompilerLocation +
			" -O3 " +
			" -lrt " +
			" -lstdc++ " +
			" -lm " +
			" -acc " +
			" -ta=nvidia " +
			" -lpgcc " +
			" -I" + CToolsDir +
			" " + filepath.Join(CToolsDir, SystemId, "libwb.a") +
			" " + filepath.Join(s.TemporaryDirectory, s.ProgramFileName) +
			" -o " + filepath.Join(s.TemporaryDirectory, s.ExecutableFileName) +
			" -DWB_USE_CUSTOM_MALLOC " +
			" -DWB_USE_COURSERA " +
			" 2>&1\n"
		revel.TRACE.Println(cmp)
		return cmp
	default:
		return "todo"
	}
}
