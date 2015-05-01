package bigcode

import (
	"path/filepath"
	. "wb/app/config"
)

// cufile -> bcfile -> o3file
// writes a shell script to generate bcfile and o3file from cufile
func MakeBigCodeByteCodeCommand(cufile string, bcfile string, o3file string) string {
	var cuflags string = `-Xclang -fcuda-is-device -femit-all-decls \
-Wno-unknown-attributes -Wno-ignored-attributes \
-Xclang -ffake-address-space-map ` +
		" -I" + filepath.Join(CUDAToolkitDirectory, "include") +
		" -I" + CToolsDir + " " +
		`-D__CUDACC__ -D__SM_32_INTRINSICS_H__ \
-D__SURFACE_INDIRECT_FUNCTIONS_H__ -DWB_USE_CUDA -DWB_USE_OPENCL \
-D__TEXTURE_INDIRECT_FUNCTIONS_H__ `

	switch OperatingSystem {
	case "Linux":
		var cmd string = "#!/bin/sh\n" +
			BigCodeClangPath + " " + cuflags +
			" -emit-llvm " +
			" -c " +
			" " + cufile + " " +
			" -o " + bcfile + " \n " +
			BigCodeOptPath +
			" -load " + BigCodeLibraryPath +
			" -remove-loodr " +
			" -O3 " +
			" " + bcfile +
			" -o " + o3file
		return cmd
	default:
		return "todo"
	}
}

func MakeBigCodeDatabaseCommand(O3file string) string {
	var bigcodeflags string = `-remove-loodr \
-scope=module \
-bigcode-noid \
-funcargfactsig \
-looplengthfactsig \
-looprepfactsig \
-arrayindexfactsig \
-array-op-rel-sig \
-callgraphfactsig \
-constantfactsig \
-structtypefactsig \
-empty-sig \
-addrspacefactsig \
-opnumfactsig \
-atomicfactsig \
-idfactsig \
-array-sum-fact-sig \
-print-sig \
`
	switch OperatingSystem {
	case "Linux":
		var cmd string = "#!/bin/sh\n" +
			BigCodeOptPath +
			" --load " + BigCodeLibraryPath +
			" " + bigcodeflags +
			" " + O3file +
			" -o /dev/null"
		return cmd
	default:
		return "todo"
	}
}
