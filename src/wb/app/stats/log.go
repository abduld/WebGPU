package stats

import (
	"log"
	"runtime"
	"time"

	"github.com/revel/revel"
)

var (
	TRACE *log.Logger
	INFO  *log.Logger
	WARN  *log.Logger
	ERROR *log.Logger
)

func doLog(file string, line int, category string, name string, v ...interface{}) {
}

func getFileLine(calldepth int) (string, int) {
	_, file, line, ok := runtime.Caller(calldepth)
	if !ok {
		file = "???"
		line = 0
	}
	return file, line
}

func Log(category, name string, val ...interface{}) {
	file, line := getFileLine(2)
	go func() {
		doLog(file, line, category, name, val)
	}()
}

func Incr(category, name string) {
	file, line := getFileLine(2)
	go func() {
		doLog(file, line, category, name, "Increment")
	}()
}

func Decr(category, name string) {
	file, line := getFileLine(2)
	go func() {
		doLog(file, line, category, name, "Decrement")
	}()
}

func LogTime(category, name string, start time.Time, end time.Time) {
	file, line := getFileLine(3)
	duration := end.Sub(start).Nanoseconds()
	doLog(file, line, category, name, duration)
}

const logFlags = log.Ldate | log.Ltime | log.Lshortfile

func InitStatsLogger() {
	TRACE = revel.TRACE
	INFO = revel.INFO
	WARN = revel.WARN
	ERROR = revel.ERROR
}
