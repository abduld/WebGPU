package util

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"time"
)

type TimeoutError struct {
	Timeout float64
}

func (p TimeoutError) Error() string {
	return fmt.Sprintf("Timeout error %v", p.Timeout)
}

func RunCommandWithTimeout(timeout time.Duration, cmd *exec.Cmd) error {
	var err error
	if err = cmd.Start(); err != nil {
		return err
	}
	errc := make(chan error, 1)
	go func() {
		errc <- cmd.Wait()
	}()
	select {
	case <-time.After(timeout):
		cmd.Process.Kill()
		err = TimeoutError{timeout.Seconds()}
	case err = <-errc:
	}
	return err
}

func RunFunctionWithTimeout(timeout time.Duration, f interface{}, xs ...interface{}) error {
	var err error
	vf := reflect.ValueOf(f)
	vxs := reflect.ValueOf(xs)
	errc := make(chan error, 1)
	go func() {
		res := vf.Call([]reflect.Value{vxs})[0].Interface()
		errc <- res.(error)
	}()
	select {
	case <-time.After(timeout):
		err = TimeoutError{timeout.Seconds()}
	case err = <-errc:
	}
	return err
}

func IsTimeoutError(e error) bool {
	switch e.(type) {
	case TimeoutError:
		return true
	}
	return false
}

func ParentDirectory(file string) string {
	return filepath.Dir(file)
}

func FileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func DoEvery(d time.Duration, f func(time.Time)) {
	go func() {
		f(time.Now())
		for x := range time.Tick(d) {
			f(x)
		}
	}()
}
