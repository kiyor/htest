/* -.-.-.-.-.-.-.-.-.-.-.-.-.-.-.-.-.-.-.-.

* File Name : example.go

* Purpose :

* Creation Date : 03-26-2016

* Last Modified : Thu Jun 16 15:29:29 2016

* Created By : Kiyor

_._._._._._._._._._._._._._._._._._._._._.*/

package htest

import ()

func ExampleConfig() []*Config {
	requestHeader := make(map[string]string)
	requestHeader["X-Debug"] = "true"

	request := Request{
		Hostname:    "a.com",
		Uri:         "/b",
		UserAgent:   "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_10_5) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/49.0.2623.87 Safari/537.36",
		Scheme:      "https",
		Method:      "GET",
		SkipTls:     true,
		Compression: false,
		Timeout:     "30s",
		Include:     []string{"template"},
		Header:      requestHeader,
	}

	requirementHeader := make(map[string][]*Factor)

	requirementHeader["X-Cache"] = append(requirementHeader["X-Cache"], &Factor{
		Obj:    "hit",
		Method: "include",
		Option: []string{"ignore case"},
	})

	requirement := Requirement{
		StatusCode: 200,
		Include:    []string{"template"},
		Header:     requirementHeader,
	}

	config := &Config{
		Hash:        "http://a.com:Mozilla",
		Request:     request,
		Requirement: requirement,
	}
	configs := []*Config{config}
	return configs

}
