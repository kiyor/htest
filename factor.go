/* -.-.-.-.-.-.-.-.-.-.-.-.-.-.-.-.-.-.-.-.

* File Name : factor.go

* Purpose :

* Creation Date : 03-26-2016

* Last Modified : Tue Apr  5 15:16:56 2016

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
