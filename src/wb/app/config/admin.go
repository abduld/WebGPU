package config

import (
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/revel/revel"
)

var (
	SendAdminEmails bool
	ECE408Students  []string
	ECE598Students  []string
	ECE408Admins    []string
	ECE598Admins    []string
	SiteAdmins      []string
)

func unionStringLists(a []string, b []string) []string {

	tmp := map[string]bool{}
	for _, admin := range a {
		if _, ok := tmp[admin]; !ok {
			tmp[admin] = true
		}
	}
	for _, admin := range b {
		if _, ok := tmp[admin]; !ok {
			tmp[admin] = true
		}
	}

	res := make([]string, 0, len(tmp))
	for k := range tmp {
		res = append(res, k)
	}
	return res
}

func lower(lst []string) []string {
	for ii, val := range lst {
		lst[ii] = strings.TrimSpace(strings.ToLower(val))
	}
	return lst
}

func readStudentFile(fileName string) []string {
	if buff, err := ioutil.ReadFile(fileName); err == nil {
		s := lower(strings.Split(string(buff), "\n"))
		res := s
		for _, elem := range s {
			res = append(res, "__2013_"+elem)
		}
		return res
	} else {
		return []string{}
	}
}

func InitAdminConfig() {

	if t, found := NestedRevelConfig.Bool("admin.send_emails"); found {
		SendAdminEmails = t
	} else {
		SendAdminEmails = true
	}

	if t, found := NestedRevelConfig.String("admin.ece408_admins"); found {
		ECE408Admins = lower(strings.Split(t, ","))
	}
	if t, found := NestedRevelConfig.String("admin.ece598_admins"); found {
		ECE598Admins = lower(strings.Split(t, ","))
	}

	SiteAdmins = unionStringLists(ECE408Admins, ECE598Admins)

	ECE408Students = readStudentFile(filepath.Join(BasePath, "ece408students.txt"))
	ECE598Students = readStudentFile(filepath.Join(BasePath, "ece598students.txt"))

	revel.TRACE.Println("Site admins")
	revel.TRACE.Println(SiteAdmins)
}
