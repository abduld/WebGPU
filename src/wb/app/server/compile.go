package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
	. "wb/app/config"
	"wb/app/stats"
	. "wb/app/util"

	"github.com/robfig/revel"
)

type SolutionData struct {
	CorrectQ bool   `json:"correctq"`
	Message  string `json:"message"`
}

type TimerElement struct {
	Id            int    `json:"id"`
	StoppedQ      bool   `json:"stopped"`
	Kind          string `json:"kind"`
	StartTime     int64  `json:"start_time"`
	EndTime       int64  `json:"end_time"`
	ElapsedTime   int64  `json:"elapsed_time"`
	StartLine     int64  `json:"start_line"`
	EndLine       int64  `json:"end_line"`
	StartFunction string `json:"start_function"`
	EndFunction   string `json:"end_function`
	StartFile     string `json:"start_file"`
	EndFile       string `json:"end_file"`
	ParentId      int64  `json:"parent_id"`
	Message       string `json:"message"`
}

type LoggerElement struct {
	Level    string `json:"level"`
	Message  string `json:"message"`
	File     string `json:"file"`
	Function string `json:"function"`
	Line     int64  `json:"line"`
	Time     int64  `json:"time"`
}

type Timer struct {
	StartTime   int64          `json:"start_time"`
	EndType     int64          `json:"end_time"`
	ElapsedTime int64          `json:"elapsed_time"`
	Elements    []TimerElement `json:"elements"`
}

type Logger struct {
	Elements []LoggerElement `json:"elements"`
}

type InternalCData struct {
	CUDAMemory     int64        `json:"cuda_memory"`
	Timer          Timer        `json:"timer"`
	Logger         Logger       `json:"logger"`
	SolutionExists bool         `json:"solution_exists"`
	Solution       SolutionData `json:"solution"`
}

type TimeoutErr struct {
	Timeout float64
}

func runOutput(timeout time.Duration, envv []string,
	stdout io.Writer, stderr io.Writer,
	dir string, argv []string) (bool, error) {

	cmd := exec.Command(argv[0], argv[1:]...)
	cmd.Dir = dir
	cmd.Env = envv
	cmd.Stdout = stdout
	cmd.Stderr = stderr

	if err := RunCommandWithTimeout(timeout, cmd); err != nil {
		return false, err
	}
	return true, nil
}

func checkSandbox(s *WorkerState) {
	s.SandboxKeyword, s.Sandboxed = IsSandboxed(s.Program)
	if s.Sandboxed {
		stats.Log("Worker", "Sandboxed", s.SandboxKeyword)
		panic("Failed in sandbox")
	}
}

func directoryExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func createTempDir(s *WorkerState) {
	runId := strings.Replace(s.RunId, "-", "", -1)
	tmpDir1, runId := runId[:5], runId[5:]
	tmpDir2, runId := runId[:5], runId[5:]
	tmp := filepath.Join(SystemTemporaryDirectory, tmpDir1, tmpDir2, "wb"+runId)
	if !directoryExists(tmp) {
		if err := os.MkdirAll(tmp, 0777); err != nil {
			stats.Log("Worker", "Error", "Failed to create temporary directory")
			panic("Failed to create tempDir")
		}
		stats.Log("Worker", "CreateTemporaryDirectory", tmp)
		s.TemporaryDirectory = tmp
		return
	}

	create := func() bool {
		idx := <-Uniq
		tmp := filepath.Join(SystemTemporaryDirectory, "wb"+fmt.Sprint(idx))
		if directoryExists(tmp) {
			return false
		}
		if err := os.MkdirAll(tmp, 0777); err != nil {
			stats.Log("Worker", "Error", "Failed to create temporary directory")
			panic("Failed to create tempDir")
		}
		stats.Log("Worker", "CreateTemporaryDirectory", tmp)
		s.TemporaryDirectory = tmp
		return true
	}
	for create() == false {
	}
}

func makeCompileCommand(s *WorkerState) string {
	switch s.Language {
	case "CUDA":
		return MakeNVCCCompileCommand(s)
	case "OpenCL":
		return MakeOpenCLCompileCommand(s)
	case "OpenACC":
		return MakeOpenACCCompileCommand(s)
	case "C++AMP":
		return MakeCPPAMPCompileCommand(s)
	default:
		return MakeNVCCCompileCommand(s)
	}
}

func createBuildFile(s *WorkerState) {
	var buildFileBaseName string
	switch OperatingSystem {
	case "Windows":
		buildFileBaseName = "build.bat"
	default:
		buildFileBaseName = "build.sh"
	}
	buildFile := filepath.Join(s.TemporaryDirectory, buildFileBaseName)
	fo, err := os.Create(buildFile)
	if err != nil {
		stats.Log("Worker", "Error", "Failed to create build file")
		panic("Failed to create build file")
	}
	defer func() {
		if fo.Close() != nil {
			stats.Log("Worker", "Error", "Failed to close build file")
			panic("Failed to close build file")
		}
	}()
	compileCommand := makeCompileCommand(s)
	if _, err = fo.Write([]byte(compileCommand)); err != nil {
		stats.Log("Worker", "Error", "Failed to write build file")
		panic("Failed to write build file")
	}
	os.Chmod(buildFile, 0777)
	s.BuildFileName = buildFile
}

func writeProgramToFile(s *WorkerState) {
	programFileName := filepath.Join(s.TemporaryDirectory, s.ProgramFileName)
	fo, err := os.Create(programFileName)
	if err != nil {
		stats.Log("Worker", "Error", "Failed to create program file")
		panic("Failed to create program file")
	}
	defer func() {
		if fo.Close() != nil {
			stats.Log("Worker", "Error", "Failed to close program file")
			panic("Failed to close program file")
		}
	}()
	if _, err = fo.Write([]byte(s.Program)); err != nil {
		stats.Log("Worker", "Error", "Failed to write program file")
		panic("Failed to write program file")
	}
	os.Chmod(programFileName, 0666)
	s.ProgramFileName = programFileName
}

func runBuildFile(s *WorkerState) {
	var stdout, stderr bytes.Buffer

	s.CompileStartTime = time.Now()
	_, err := runOutput(CompileTimeout,
		nil, /* env */
		&stdout,
		&stderr,
		filepath.Dir(s.BuildFileName),
		[]string{s.BuildFileName},
	)
	s.CompileEndTime = time.Now()

	stats.LogTime("Worker", "CompileTime", s.CompileStartTime, s.CompileEndTime)

	s.CompileStdout = string(stdout.Bytes())
	s.CompileStderr = string(stderr.Bytes())

	if err != nil {
		switch err.(type) {
		case TimeoutError:
			s.TimeoutError = true
			s.TimeoutValue = err.(TimeoutError).Timeout
			stats.Incr("Worker", "RunTimeout")
			panic("Failed to compile program")
			return
		}
	}
	if err != nil ||
		!FileExists(filepath.Join(s.TemporaryDirectory, s.ExecutableFileName)) {
		s.CompilationFailed = true
		stats.Incr("Worker", "CompilationFailed")
		panic("Failed to compile program")
	}
}

const seperator = "==$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$"

func runProgram(s *WorkerState) {
	var stdout, stderr bytes.Buffer

	stats.Incr("Worker", "Run")

	conf := s.MachineProblemConfig

	outputFileExtension := func() string {
		if conf.OutputType == "image" {
			return "ppm"
		} else {
			return "raw"
		}
	}

	getInputs := func() []string {
		if s.DatasetId < 0 {
			return []string{"none"}
		} else {
			datasetConfig := conf.Datasets[s.DatasetId]
			inputs := make([]string, len(datasetConfig.Input))
			for i, input := range datasetConfig.Input {
				inputs[i] = filepath.Join(s.MachineProblemDirectory, "data", input)
			}
			return inputs
		}
	}

	getOutput := func() string {
		if s.DatasetId < 0 {
			return "none"
		} else {
			datasetConfig := conf.Datasets[s.DatasetId]
			output := datasetConfig.Output
			return filepath.Join(s.MachineProblemDirectory, "data", output)
		}
	}

	s.UserOutput = filepath.Join(s.TemporaryDirectory, "output."+outputFileExtension())

	runCommand := []string{
		filepath.Join(s.TemporaryDirectory, s.ExecutableFileName),
		"-i ", strings.Join(getInputs(), ","),
		//"-o ", s.UserOutput,
		"-e ", getOutput(),
		"-t ", conf.OutputType,
	}

	s.RunStartTime = time.Now()
	_, err := runOutput(RuntimeTimeout,
		nil, /* env */
		&stdout,
		&stderr,
		filepath.Dir(s.BuildFileName),
		runCommand,
	)
	s.RunEndTime = time.Now()

	stats.LogTime("Worker", "RunTime", s.RunStartTime, s.RunEndTime)

	s.RunStdout = string(stdout.Bytes())
	s.RunStderr = string(stderr.Bytes())

	removeInternalData := func(s string) string {
		ss := strings.Split(s, seperator)
		return ss[0]
	}

	switch err.(type) {
	case TimeoutError:
		revel.TRACE.Println("Terminated....")
		s.TimeoutError = true
		s.TimeoutValue = err.(TimeoutError).Timeout
		stats.Incr("Worker", "RunTimeout")
		panic("Failed to run program")
	}

	if strings.Contains(s.RunStderr, "<<SANDBOXED>>") {
		s.Sandboxed = true
		s.SandboxKeyword = "Program sandboxed because of use of " +
			removeInternalData(strings.TrimPrefix(s.RunStdout, "<<SANDBOXED>>::")) +
			" keyword."
		stats.Log("Worker", "RunSandboxed", s.SandboxKeyword)
		panic("Failed to run program")
	} else if strings.Contains(s.RunStderr, "<<MEMORY>>") {
		s.RunFailed = true
		s.RunStdout = ""
		s.RunStderr = "Program teminated because it is allocating too much memory."
		stats.Incr("Worker", "MemoryLimit")
		panic("Failed to run program")
	} else if err != nil {
		s.RunFailed = true
		s.RunStdout = removeInternalData(s.RunStdout)
		stats.Incr("Worker", "RunFailed")
		panic("Failed to run program")
	}

	ss := strings.Split(s.RunStdout, seperator)
	s.RunStdout = ss[0]

	var wbData InternalCData

	normalizeString := func(s string) string {
		res := s
		if !utf8.ValidString(s) {
			v := make([]rune, 0, len(s))
			for i, r := range s {
				if r == utf8.RuneError {
					if _, size := utf8.DecodeRuneInString(s[i:]); size == 1 {
						continue
					}
				}
				v = append(v, r)
			}
			res = string(v)
		}
		return res
	}

	err = json.Unmarshal([]byte(normalizeString(ss[1])), &wbData)
	if err != nil {
		stats.Incr("Worker", "InternalDataReadError")
		s.RunFailed = true
		s.RunStdout = s.RunStdout
		s.RunStderr = "Failed to read program output. Make sure you do not have special characters in your code."
		stats.Log("Worker", "Error", "Failed to read internal data  "+fmt.Sprint(err))
		return
	}

	if s.DatasetId < 0 {
		wbData.SolutionExists = false
		wbData.Solution.CorrectQ = true
		wbData.Solution.Message = "No solution expected."
	}

	s.SolutionCorrect = wbData.Solution.CorrectQ
	s.SolutionMessage = wbData.Solution.Message
	s.InternalCData = wbData

	if !s.SolutionCorrect {
		stats.Log("Worker", "IncorrectSolution", 1)
	} else {
		stats.Log("Worker", "CorrectSolution", 1)
	}
}

func readUserOutput(s *WorkerState) {
	path := s.UserOutput
	data, err := ioutil.ReadFile(path)
	if err == nil {
		s.UserOutput = "Cannot read user output"
	} else {
		s.UserOutput = string(data)
	}
}

func readMachineProblemConfig(s *WorkerState) {
	if conf, err := ReadMachineProblemConfig(s.MachineProblemNumber); err == nil {
		s.MachineProblemConfig = conf
	} else {
		panic("Cannot read machine problem config")
	}
}

func compile(req *WorkerRequest, datasetId int) (s *WorkerState) {

	defer func() {
		if r := recover(); r != nil {
			return
		}
	}()

	var programOutputFile string

	if req.Language == "CUDA" {
		programOutputFile = "program.cu"
	} else {
		programOutputFile = "program.cpp"
	}

	s = &WorkerState{
		MachineProblemNumber: req.MachineProblemNumber,
		MachineProblemId:     req.MachineProblemId,
		DatasetId:            datasetId,
		RunId:                req.RunId,
		Program:              req.Program,
		ProgramFileName:      programOutputFile,
		ExecutableFileName:   "program" + SystemExecutableExtension,
		TemporaryDirectory:   SystemTemporaryDirectory,
		RunFailed:            false,
		TimeoutError:         false,
		CompilationFailed:    false,
		Sandboxed:            false,
		RequestStartTime:     req.RequestStartTime,
		SolutionCorrect:      false,
		Language:             req.Language,
		ForGrading:           req.ForGrading,
		MachineProblemDirectory: filepath.Join(MPFileDirectory,
			strconv.Itoa(req.MachineProblemNumber)),
	}

	readMachineProblemConfig(s)
	checkSandbox(s)
	createTempDir(s)
	createBuildFile(s)
	writeProgramToFile(s)
	runBuildFile(s)

	return
}

func run(s *WorkerState) {

	if s.CompilationFailed || s.Sandboxed || s.TimeoutError {
		if s.CompilationFailed {
			stats.TRACE.Println("Compilation failed...")
		}
		if s.Sandboxed {
			stats.TRACE.Println("Sandboxed...")
		}
		if s.TimeoutError {
			stats.TRACE.Println("Timeout error...")
		}
		return
	}

	defer func() {
		if r := recover(); r != nil {
			return
		}
	}()
	runProgram(s)
	//readUserOutput(s)
}

func CompileAndRun(req *WorkerRequest) (res []WorkerState) {

	stats.Incr("Worker", "Compilations")

	conf, _ := ReadMachineProblemConfig(req.MachineProblemNumber)

	stats.Log("Worker", "CompileMP", "MP"+
		strconv.Itoa(req.MachineProblemNumber))
	stats.Log("Worker", "DatasetRun", "MP"+
		strconv.Itoa(req.MachineProblemNumber)+"::"+
		strconv.Itoa(req.DatasetId))

	if req.DatasetId == -1 && len(conf.Datasets) != 0 {
		res = make([]WorkerState, len(conf.Datasets))
		s := compile(req, 0)
		s.OnAllDatasets = true

		for i := range conf.Datasets {
			res[i] = *s
			res[i].DatasetId = i
			run(&res[i])
		}

		/*
			if directoryExists(s.TemporaryDirectory) {
				go func() {
					os.RemoveAll(s.TemporaryDirectory)
				}()
			}
		*/
	} else {
		res = make([]WorkerState, 1)
		s := compile(req, req.DatasetId)
		if req.DatasetId == -1 {
			s.OnAllDatasets = true
		} else {
			s.OnAllDatasets = false
		}
		run(s)
		res[0] = *s

		/*
			if directoryExists(s.TemporaryDirectory) {
				go func() {
					os.RemoveAll(s.TemporaryDirectory)
				}()
			}
		*/
	}
	return
}
