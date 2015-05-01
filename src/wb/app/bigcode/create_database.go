package bigcode

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"sync"
	. "wb/app/config"
	"wb/app/models"

	"github.com/revel/revel"
)

type workingState struct {
	mpNumber         int
	program          models.Program
	bcBuildDirectory string
	dbBuildDirectory string
	outputDirectory  string
	bcPath           string
	o3Path           string
	err              error
}

func directoryExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func buildBytecode(state workingState) error {

	var stdout, stderr bytes.Buffer
	program := state.program
	outputDirectory := state.outputDirectory
	bcBuildDirectory := state.bcBuildDirectory

	idStr := strconv.Itoa(int(program.Id))

	cuPath := filepath.Join(outputDirectory, idStr+".cu")
	bcPath := filepath.Join(outputDirectory, idStr+".bc")
	o3Path := filepath.Join(outputDirectory, idStr+".O3.bc")

	if err := ioutil.WriteFile(cuPath, []byte(program.Text), 0777); err != nil {
		revel.TRACE.Println(err)
		return err
	}

	bcCommand := MakeBigCodeByteCodeCommand(cuPath, bcPath, o3Path)
	bcBuildFileName := filepath.Join(bcBuildDirectory, idStr+"_bc_build.sh")
	if err := ioutil.WriteFile(bcBuildFileName, []byte(bcCommand), 0777); err != nil {
		revel.TRACE.Println(err)
		return err
	}
	//revel.TRACE.Println(bcBuildFileName)
	cmd := exec.Command(bcBuildFileName)
	cmd.Dir = bcBuildDirectory
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		revel.TRACE.Println("stdout = ", string(stdout.Bytes()))
		revel.TRACE.Println("stderr = ", string(stderr.Bytes()))
		return err
	}
	return nil
}

func buildSignature(state workingState) (string, error) {

	var stdout, stderr bytes.Buffer

	program := state.program

	dbBuildDirectory := state.dbBuildDirectory
	outputDirectory := state.outputDirectory

	idStr := strconv.Itoa(int(program.Id))

	o3Path := filepath.Join(outputDirectory, idStr+".O3.bc")

	dbCommand := MakeBigCodeDatabaseCommand(o3Path)
	dbBuildFileName := filepath.Join(dbBuildDirectory, idStr+"_db_build.sh")
	if err := ioutil.WriteFile(dbBuildFileName, []byte(dbCommand), 0777); err != nil {
		revel.TRACE.Println(err)
		return "", err
	}

	//revel.TRACE.Println(dbBuildFileName)
	cmd := exec.Command(dbBuildFileName)
	cmd.Dir = dbBuildDirectory
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		revel.TRACE.Println(err)
		return "", err
	}
	return string(stdout.Bytes()), nil
}

func makeBigcodeSignature(state workingState) (string, error) {
	if err := buildBytecode(state); err != nil {
		revel.TRACE.Println(err)
		return "", err
	}
	return buildSignature(state)
}

func CreateProgramSignature(mpNumber int, program models.Program) (string, error) {
	tmpDir := filepath.Join(SystemTemporaryDirectory, "bigcode", strconv.Itoa(mpNumber))
	outputDirectory := filepath.Join(tmpDir, "output")
	if err := os.MkdirAll(outputDirectory, 0700); err != nil {
		revel.TRACE.Println(err)
		return "", err
	}

	bcBuildDirectory := filepath.Join(tmpDir, "bcbuild")
	if err := os.MkdirAll(bcBuildDirectory, 0700); err != nil {
		revel.TRACE.Println(err)
		return "", err
	}

	dbBuildDirectory := filepath.Join(tmpDir, "dbbuild")
	if err := os.MkdirAll(dbBuildDirectory, 0700); err != nil {
		revel.TRACE.Println(err)
		return "", err
	}

	input := workingState{
		mpNumber:         mpNumber,
		outputDirectory:  outputDirectory,
		bcBuildDirectory: bcBuildDirectory,
		dbBuildDirectory: dbBuildDirectory,
		program:          program,
	}
	return makeBigcodeSignature(input)
}
func CreateDatabase(mpNumber int, progs []models.Program) (BigCodeDatabase, error) {
	var database BigCodeDatabase
	database.MPNumber = mpNumber

	tmpDir := filepath.Join(SystemTemporaryDirectory, "bigcode", strconv.Itoa(mpNumber))
	//os.RemoveAll(tmpDir)

	if err := os.MkdirAll(tmpDir, 0700); err != nil {
		return database, err
	}

	outputDirectory := filepath.Join(tmpDir, "output")
	if !directoryExists(outputDirectory) {
		if err := os.MkdirAll(outputDirectory, 0777); err != nil {
			return database, err
		}
	}
	bcBuildDirectory := filepath.Join(tmpDir, "bcbuild")
	if !directoryExists(bcBuildDirectory) {
		if err := os.MkdirAll(bcBuildDirectory, 0777); err != nil {
			return database, err
		}
	}
	dbBuildDirectory := filepath.Join(tmpDir, "dbbuild")
	if !directoryExists(dbBuildDirectory) {
		if err := os.MkdirAll(dbBuildDirectory, 0777); err != nil {
			return database, err
		}
	}

	var wg sync.WaitGroup
	wg.Add(BigCodeNumberOfBatches)

	step := (len(progs) + BigCodeNumberOfBatches - 1) / BigCodeNumberOfBatches
	for ii := 0; ii < BigCodeNumberOfBatches; ii++ {
		go func(ii int) {
			defer wg.Done()
			start := ii * step
			end := (ii + 1) * step
			if end > len(progs) {
				end = len(progs)
			}
			if start >= end {
				return
			}
			for _, program := range progs[start:end] {
				if sig, err := CreateProgramSignature(mpNumber, program); err == nil && sig != "" {
					addSignature(mpNumber, program, sig)
				}
			}

		}(ii)
	}
	wg.Wait()

	database.Entries = []BigCodeDatabaseEntry{}
	for _, prog := range progs {
		if sig, _ := getSignature(mpNumber, prog.Id); sig != "" {
			database.Entries = append(database.Entries, BigCodeDatabaseEntry{
				ProgramId: prog.Id,
				Code:      prog.Text,
				Signature: sig,
			})
		}
	}

	dbJson, _ := json.MarshalIndent(database, "", "  ")

	outputFile := filepath.Join(MPFileDirectory, strconv.Itoa(mpNumber), "bigcode_database.json")
	os.Remove(outputFile)
	ioutil.WriteFile(outputFile, []byte(dbJson), 0666)

	return database, nil
}
