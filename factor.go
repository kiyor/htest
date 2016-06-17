/* -.-.-.-.-.-.-.-.-.-.-.-.-.-.-.-.-.-.-.-.

* File Name : factor.go

* Purpose :

* Creation Date : 03-26-2016

* Last Modified : Thu Jun 16 12:37:38 2016

* Created By : Kiyor

_._._._._._._._._._._._._._._._._._._._._.*/

package htest

import (
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strings"
)

var supportMethodList = []string{"exist", "not exist", "include", "not include", "match", "not match", "regex match", "regex not match", "show", "show if exist"}

type Factor struct {
	sub    string
	Obj    string
	Method string
	Option []string
}

func fixHeader(value string) string {
	return http.CanonicalHeaderKey(value)
}

func SupportMethod() string {
	var out string
	for _, v := range supportMethodList {
		out += fmt.Sprintf("- %s\n", v)
	}
	return out
}

func cut(left, right int, val string) string {
	if len(val)-right >= left {
		return val[left : len(val)-right]
	}
	return val
}

func rmQuerystring(name, url string) string {
	p := strings.Split(url, "?")
	if len(p) == 1 {
		return url
	}
	newurl := p[0] + "?"
	for _, v := range strings.Split(p[1], "&") {
		kv := strings.Split(v, "=")
		if kv[0] != name {
			newurl += v + "&"
		}
	}
	newurl = strings.TrimRight(newurl, "&")

	return newurl
}

func (f *Factor) Pass() (string, bool, error) {
	sub, obj := f.sub, f.Obj
	for _, o := range f.Option {
		if o == "ignore case" {
			sub = strings.ToLower(f.sub)
			obj = strings.ToLower(f.Obj)
		}
	}
	switch f.Method {
	case "exist":
		return "", sub != "nil", nil
	case "not exist":
		return "", sub == "nil", nil
	case "include":
		return "", strings.Contains(sub, obj), nil
	case "not include":
		return "", !strings.Contains(sub, obj), nil
	case "match":
		return "", sub == obj, nil
	case "not match":
		return "", sub != obj, nil
	case "show":
		if sub == "nil" {
			return sub, false, nil
		}
		return sub, true, nil
	case "show if exist":
		return sub, true, nil
	case "regex match":
		re, err := regexp.Compile(obj)
		if err != nil {
			return "", false, err
		}
		return "", re.MatchString(sub), nil
	case "regex not match":
		re, err := regexp.Compile(obj)
		if err != nil {
			return "", false, err
		}
		return "", !re.MatchString(sub), nil
	}

	return "", false, errors.New("method not found: " + f.Method)
}
