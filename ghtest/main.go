/* -.-.-.-.-.-.-.-.-.-.-.-.-.-.-.-.-.-.-.-.

* File Name : main.go

* Purpose :

* Creation Date : 03-26-2016

* Last Modified : Mon May  8 12:17:05 2017

* Created By : Kiyor

_._._._._._._._._._._._._._._._._._._._._.*/

package main

import (
	"flag"
	"fmt"
	"github.com/kiyor/golib"
	"github.com/kiyor/htest"
	"gopkg.in/yaml.v2"
	"log"
	"net/url"
	"os"
	"runtime"
	"strings"
)

const (
	RETURN_GOOD = iota
	RETURN_BAD
	RETURN_CONFIG_ERR
)

var (
	flagConfig        *string = flag.String("c", "./config", "config file or path")
	flagTemplate      *string = flag.String("t", "./template", "template file or path")
	flagExampleConfig *bool   = flag.Bool("example", false, "generate example config")
	flagIp            *string = flag.String("ip", "", "testing ip")
	flagVerbose       *bool   = flag.Bool("vv", false, "verbose output")
	flagRaw           *bool   = flag.Bool("raw", false, "raw response output")
	flagMethod        *bool   = flag.Bool("method", false, "check support compare method")
	flagTimeout       *string = flag.String("timeout", "", "overwrite timeout value")
	flagCheckOnly     *bool   = flag.Bool("check", false, "check config file only")
	flagNewConfig     *string = flag.String("new", "http://a.com/b", "create new config")
	flagCurl          *bool   = flag.Bool("curl", false, "output curl command")
	flagVersion       *bool   = flag.Bool("v", false, "print version and exist")

	VER       = "1.0"
	buildtime string
)

func init() {
	flag.Parse()
	if *flagVersion {
		fmt.Printf("%v.%v", VER, buildtime)
		os.Exit(0)
	}
	htest.Logger = golib.NewLogger(&golib.LogOptions{
		Name:      "htest",
		ShowErr:   true,
		ShowDebug: *flagVerbose,
		ShowColor: true,
	})
	htest.Verbose = *flagVerbose
	htest.RawResp = *flagRaw
	htest.ShowCurl = *flagCurl

	if *flagMethod {
		fmt.Println(htest.SupportMethod())
		os.Exit(0)
	}
	if *flagExampleConfig {
		configs := htest.ExampleConfig()
		d, _ := yaml.Marshal(configs)
		fmt.Println(string(d))
		os.Exit(0)
	}
	if *flagNewConfig != "http://a.com/b" {
		configs := htest.ExampleConfig()
		u, _ := url.Parse(*flagNewConfig)
		configs[0].Hash = *flagNewConfig
		configs[0].Request.Scheme = u.Scheme
		configs[0].Request.Uri = u.RequestURI()
		configs[0].Request.Hostname = u.Host
		d, _ := yaml.Marshal(configs)
		fmt.Println(string(d))
		os.Exit(0)
	}
	if *flagCheckOnly {
		htest.VerifyYaml(*flagTemplate, true)
		htest.VerifyYaml(*flagConfig, false)
		os.Exit(0)
	}

	htest.OverTimeout = *flagTimeout
	runtime.GOMAXPROCS(runtime.NumCPU())

	htest.LoadTemplate(*flagTemplate)

	log.SetFlags(log.Lshortfile)
}

func main() {
	results := make(chan *htest.Result)
	exit := RETURN_GOOD
	go func() {
		for {
			select {
			case r := <-results:
				fmt.Println(r.String())
				if !r.AllPass() {
					exit = RETURN_BAD
				}
			}
		}
	}()
	ips := strings.Split(*flagIp, ",")
	err := htest.DoCheck(*flagConfig, results, ips...)
	if err != nil {
		os.Exit(RETURN_CONFIG_ERR)
	}
	os.Exit(exit)
}
