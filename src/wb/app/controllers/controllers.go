package controllers

import (
	"net/http"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"time"
	. "wb/app/config"

	"github.com/revel/revel"
)

var (
	createBigCodeDatabaseMutex *sync.Mutex
)

func InitControllers() {

	revel.InterceptMethod(CheckCourseraMode, revel.BEFORE)
	revel.InterceptMethod(PublicApplication.AddUser, revel.BEFORE)
	revel.InterceptMethod(CheckWorker, revel.BEFORE)
	revel.InterceptFunc(CheckUser, revel.BEFORE, &SecuredApplication{})
	if CourseraMode {
		revel.InterceptFunc(CheckUser, revel.BEFORE, &CourseraApplication{})
	}

	revel.TemplateFuncs["rfc3339"] = func(t time.Time) string {
		return t.Format(time.RFC3339)
	}
	revel.TemplateFuncs["lower"] = func(s string) string {
		return strings.ToLower(s)
	}
	revel.TemplateFuncs["plus"] = func(x int, y int) int {
		return x + y
	}
	revel.TemplateFuncs["shorten"] = func(s string) string {
		if len(s) > 60 {
			bs := []byte(s)
			return string(bs[:60]) + "..."
		} else {
			return s
		}
	}
	revel.TemplateFuncs["percentageToInt"] = func(f float32) int {
		return int(f * 100)
	}
	revel.TemplateFuncs["splitComma"] = func(str string) []string {
		return strings.Split(str, ",")
	}
	revel.TemplateFuncs["positiveQ"] = func(i int) bool {
		return i > 0
	}
	revel.TemplateFuncs["notEmpty"] = func(xs interface{}) bool {
		vxs := reflect.ValueOf(xs)
		return vxs.Len() > 0
	}
	revel.TemplateFuncs["loggerClass"] = func(s string) (class string) {
		switch s {
		case "Fatal":
		case "Error":
			class = "danger"
		case "Warn":
			class = "warning"
		case "Info":
		case "Debug":
			class = "active"
		default:
			class = ""
		}
		return
	}
	revel.TemplateFuncs["milliSeconds"] = func(t int64) float64 {
		return float64(t) / 1000000.0
	}

	MPFileDirectory = filepath.Join(revel.BasePath, "mp")
	createBigCodeDatabaseMutex = &sync.Mutex{}
	if CourseraMode {
		InitCourseraController()
	}
	if IsMaster {
		InitAdminController()
	}
}

func ResetControllers() {
	NoticeCache = nil
	InitAdminController()
}

type RenderHtmlResult struct {
	html string
}

func (r RenderHtmlResult) Apply(req *revel.Request, resp *revel.Response) {
	resp.WriteHeader(http.StatusOK, "text/html")
	resp.Out.Write([]byte(r.html))
}

func (c PublicApplication) RenderHtml(o string) revel.Result {
	return RenderHtmlResult{o}
}

func JSTime(s string) (time.Time, error) {
	return time.Parse(time.RFC1123, s)
}
