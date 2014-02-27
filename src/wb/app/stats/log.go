package stats

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"runtime"
	"time"
	. "wb/app/config"

	"github.com/abduld/influxdbc"
	"github.com/robfig/revel"
)

type Logger struct {
	Type        string
	Prefix      string
	IOLogger    *log.Logger
	RedisLogger *RedisConnection
}

var (
	TRACE *Logger
	INFO  *Logger
	WARN  *Logger
	ERROR *Logger
)

type RedisLog struct {
	Category string    `json:"category"`
	Name     string    `json:"name"`
	File     string    `json:"file"`
	Line     int       `json:"line"`
	Time     time.Time `json:"time"`
	Message  string    `json:"message"`
}

func doLog(file string, line int, category string, name string, v ...interface{}) {
	if EnableRedisStore {
		pkt := RedisLog{
			Category: category,
			Name:     name,
			File:     file,
			Line:     line,
			Time:     time.Now(),
			Message:  fmt.Sprint(v...),
		}

		if js, err := json.Marshal(pkt); err == nil {
			RedisPool.Set(category, string(js))
		}
	}
	if EnableInfluxStore {
		go func() {
			defer func() {
				if r := recover(); r != nil {
					return
				}
			}()
			sval := fmt.Sprint(v...)
			series := influxdbc.NewSeries("role", "category", "state", "time", "value")
			series.AddPoint(ServerRole, category, name, time.Now().Format(time.RFC3339), sval)
			InfluxClient.WriteSeries([]influxdbc.Series{*series})
		}()
	}
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
		if EnableRedisStore {
			RedisPool.Incr(category + "::" + name)
		}
	}()
}

func Decr(category, name string) {
	file, line := getFileLine(2)
	go func() {
		doLog(file, line, category, name, "Decrement")
		if EnableRedisStore {
			RedisPool.Decr(category + "::" + name)
		}
	}()
}

func LogTime(category, name string, start time.Time, end time.Time) {
	file, line := getFileLine(3)
	duration := end.Sub(start).Nanoseconds()
	doLog(file, line, category, name, duration)
}

func (logger *Logger) Println(v ...interface{}) {
	s := fmt.Sprint(v...)
	switch logger.Type {
	case "Redis":
		file, line := getFileLine(2)
		pkt := RedisLog{
			Category: "Log",
			Name:     logger.Prefix,
			File:     file,
			Line:     line,
			Time:     time.Now(),
			Message:  s,
		}

		if js, err := json.Marshal(pkt); err == nil {
			logger.RedisLogger.Append("Log", string(js))
		}
	}
	logger.IOLogger.Println("=> " + s)
}

const logFlags = log.Ldate | log.Ltime | log.Lshortfile

func newIOLogger(prefix string, wr io.Writer) *Logger {
	logger := Logger{
		Type:     "IO",
		Prefix:   prefix,
		IOLogger: log.New(wr, prefix+" :: ", logFlags),
	}
	return &logger
}

func newRedisLogger(prefix string) *Logger {
	var wr io.Writer
	file, err := os.OpenFile("log."+prefix+".output", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		revel.ERROR.Println("Failed to open log file", prefix, ":", err)
		wr = os.Stderr
	} else {
		wr = file
	}
	conn, err := CreateRedisConnection()
	logger := Logger{
		Type:        "Redis",
		Prefix:      prefix,
		IOLogger:    log.New(wr, prefix+" :: ", logFlags),
		RedisLogger: &RedisConnection{conn},
	}
	if err != nil {
		revel.WARN.Println("Cannot make new redis connection... setting logger to IO")
		logger.Type = "IO"
		logger.RedisLogger = nil
	}
	return &logger
}

func getLogger(prefix string, output string) *Logger {
	var logger *Logger
	switch output {
	case "stdout":
		logger = newIOLogger(prefix, os.Stdout)
	case "stderr":
		logger = newIOLogger(prefix, os.Stderr)
	case "redis":
		if !EnableRedisStore {
			revel.WARN.Println("Trying to use a redis log store even " +
				"though it's disabled. Using stdout as log output.")
			logger = newIOLogger(prefix, os.Stdout)
		} else {
			logger = newRedisLogger(prefix)
		}
	default:
		if output == "off" {
			output = os.DevNull
		}
		file, err := os.OpenFile(output, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			log.Fatalln("Failed to open log file", output, ":", err)
		}
		logger = newIOLogger(output, file)
	}

	return logger
}

func InitStatsLogger() {
	TRACE = getLogger("Trace", TraceLogOutput)
	INFO = getLogger("Info", InfoLogOutput)
	WARN = getLogger("Warn", WarnLogOutput)
	ERROR = getLogger("Error", ErrorLogOutput)
}
