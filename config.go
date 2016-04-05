/* -.-.-.-.-.-.-.-.-.-.-.-.-.-.-.-.-.-.-.-.

* File Name : config.go

* Purpose :

* Creation Date : 03-25-2016

* Last Modified : Tue Apr  5 16:09:58 2016

* Created By : Kiyor

_._._._._._._._._._._._._._._._._._._._._.*/

package htest

import (
	"encoding/json"
	"fmt"
	"github.com/wsxiaoys/terminal/color"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"
)

var (
	Verbose     bool
	OverTimeout string
	MaxConn     = 10
	TemplateMap = make(map[string]*Config)
	RawResp     bool
	ShowCurl    bool
)

func toJson(i interface{}) string {
	b, _ := json.MarshalIndent(i, "", "  ")
	return string(b)
}

type Config struct {
	file        string
	Hash        string `,omitempty`
	Request     Request
	Requirement Requirement
}

type Request struct {
	client      *http.Client
	Hostname    string
	Uri         string
	Method      string
	Scheme      string
	UserAgent   string
	testIp      string
	KeepAlive   bool
	SkipTls     bool
	Compression bool
	Timeout     string
	Include     []string `,omitempty`
	Header      map[string]string
}

func (r *Request) toUrl() string {
	if len(r.testIp) == 0 {
		r.testIp = r.Hostname
	}
	return fmt.Sprintf("%s://%s%s", r.Scheme, r.testIp, r.Uri)
}
func (c *Config) Title() string {
	if len(c.Request.testIp) == 0 {
		c.Request.testIp = c.Request.Hostname
	}
	id := c.file
	if len(c.Hash) > 0 {
		id += ":" + c.Hash
	}
	return fmt.Sprintf("%s %s %s://%s%s", c.Request.testIp, id, c.Request.Scheme, c.Request.Hostname, c.Request.Uri)
}

type Requirement struct {
	StatusCode int
	Include    []string `,omitempty`
	Header     map[string][]*Factor
}

type Result struct {
	Title    string
	Pass     []string
	NotPass  []string
	Error    []string
	Duration time.Duration
	config   *Config
	rawResp  *http.Response
}

func LoadTemplate(path string) {
	filepath.Walk(path, func(p string, s os.FileInfo, e error) error {
		if e != nil {
			fmt.Println("config not founc")
			os.Exit(1)
		}
		l, _ := os.Readlink(p)
		if len(l) > 0 {
			p = l
			s, _ = os.Lstat(p)
			if s.IsDir() {
				LoadTemplate(p)
			}
		}
		if !s.IsDir() && len(l) == 0 {
			loadTemplate(p)
		}
		return nil
	})
}
func loadTemplate(file string) {
	var config Config
	var configs []*Config

	load := func(c *Config) {
		if len(c.Hash) == 0 {
			Logger.Warning("Template will only load the config with hash", c.Title())
		} else {
			if _, ok := TemplateMap[c.Hash]; ok {
				Logger.Warning("Template already loaded, ignore", c.Title())
			} else {
				TemplateMap[c.Hash] = c
			}
		}
	}

	b, err := ioutil.ReadFile(file)
	if err != nil {
		Logger.Error(err.Error())
	}
	err1 := yaml.Unmarshal(b, &config)
	err2 := yaml.Unmarshal(b, &configs)
	if err1 != nil && err2 != nil {
		Logger.Error(file, err1.Error(), err2.Error())
		return
	}

	// config is a single config
	if err1 == nil {
		load(&config)
	}
	// config is a list of config
	if err2 == nil {
		for _, c := range configs {
			load(c)
		}
	}
}

func VerifyYaml(path string, checkTemplate bool) {
	filepath.Walk(path, func(p string, s os.FileInfo, e error) error {
		if e != nil {
			fmt.Println("config not found")
			os.Exit(1)
		}
		l, _ := os.Readlink(p)
		if len(l) > 0 {
			p = l
			s, _ = os.Lstat(p)
			if s.IsDir() {
				VerifyYaml(p, checkTemplate)
			}
		}
		if !s.IsDir() && len(l) == 0 {
			verifyYaml(p, checkTemplate)
		}
		return nil
	})
}
func verifyYaml(file string, checkTemplate bool) {
	var config Config
	var configs []*Config

	check := func(c *Config) {
		for _, fs := range c.Requirement.Header {
			for _, f := range fs {
				p := false
				for _, method := range supportMethodList {
					if method == f.Method {
						p = true
						break
					}
				}
				if !p {
					Logger.Critical(f.Method, "not support in", file)
				}
			}
		}
	}

	b, err := ioutil.ReadFile(file)
	if err != nil {
		if err != nil {
			Logger.Error(err.Error())
		}
	}
	err1 := yaml.Unmarshal(b, &config)
	err2 := yaml.Unmarshal(b, &configs)
	if err1 != nil && err2 != nil {
		Logger.Error(err1.Error(), err2.Error())
		return
	}

	// config is a single config
	if err1 == nil {
		check(&config)
	}
	// config is a list of config
	if err2 == nil {
		for _, c := range configs {
			check(c)
		}
	}

}

func DoCheck(path string, results chan *Result, ips ...string) {
	configChan := make(chan *Config)
	// this wg for check, not for result output
	wg := sync.WaitGroup{}
	go Verifier(configChan, results, &wg)

	filepath.Walk(path, func(p string, s os.FileInfo, _ error) error {
		l, _ := os.Readlink(p)
		if len(l) > 0 {
			p = l
			s, _ = os.Lstat(p)
			if s.IsDir() {
				DoCheck(p, results, ips...)
			}
		}
		if !s.IsDir() && len(l) == 0 {
			doCheck(p, configChan, results, &wg, ips...)
		}

		return nil
	})
	wg.Wait()
	time.Sleep(1 * time.Second)
}

func doCheck(file string, configChan chan *Config, results chan *Result, wg *sync.WaitGroup, ips ...string) {
	var config Config
	var configs []*Config

	check := func(c *Config) {
		if len(ips) == 0 {
			Logger.Notice("put", c.Title(), "to queue")
			if c.Request.Scheme == "both" {
				c1 := *c
				c2 := *c
				wg.Add(2)
				c1.Request.Scheme = "http"
				c2.Request.Scheme = "https"
				configChan <- &c1
				configChan <- &c2
			} else {
				wg.Add(1)
				configChan <- c
			}
		} else {
			for _, ip := range ips {
				c.Request.testIp = ip
				Logger.Notice("put", c.Title(), "to queue")
				if c.Request.Scheme == "both" {
					c1 := *c
					c2 := *c
					wg.Add(2)
					c1.Request.Scheme = "http"
					c2.Request.Scheme = "https"
					configChan <- &c1
					configChan <- &c2
				} else {
					wg.Add(1)
					configChan <- c
				}
			}
		}
	}

	b, err := ioutil.ReadFile(file)
	Logger.Notice("loading file", file)
	if err != nil {
		Logger.Error("file not exist", file)
	}
	err1 := yaml.Unmarshal(b, &config)
	err2 := yaml.Unmarshal(b, &configs)

	if err1 != nil && err2 != nil {
		Logger.Error(file, err1.Error(), err2.Error())
		return
	}

	configInherit := func(config Config) *Config {
		requestHeader := config.Request.Header
		newRequestHeader := make(map[string]string)
		for _, in := range config.Request.Include {
			if template, ok := TemplateMap[in]; ok {
				for k, v := range template.Request.Header {
					newRequestHeader[k] = v
				}
			}
		}
		for k, v := range requestHeader {
			newRequestHeader[k] = v
		}
		config.Request.Header = newRequestHeader

		// store requirementheader
		requirementHeader := config.Requirement.Header
		newRequirementHeader := make(map[string][]*Factor)

		// include from template
		for _, in := range config.Requirement.Include {
			if template, ok := TemplateMap[in]; ok {
				for k, v := range template.Requirement.Header {
					newRequirementHeader[k] = v
				}
			}
		}
		for k, v := range requirementHeader {
			for _, v1 := range v {
				newRequirementHeader[k] = append(newRequirementHeader[k], v1)
			}
		}

		config.Requirement.Header = newRequirementHeader

		return &config
	}

	// config is a single config
	if err1 == nil {
		config.file = file
		c := configInherit(config)
		check(c)
	}
	if err2 == nil {
		for _, config := range configs {
			config.file = file
			c := configInherit(*config)
			check(c)
		}

	}
}

func (c *Config) Do() (*http.Response, error) {
	client := &http.Client{
		Transport: NewHTTransport(c),
	}
	var err error
	if len(OverTimeout) > 0 {
		c.Request.Timeout = OverTimeout
	}
	client.Timeout, err = time.ParseDuration(c.Request.Timeout)
	if err != nil {
		client.Timeout = time.Duration(10 * time.Second)
	}

	req, err := http.NewRequest(c.Request.Method, c.Request.toUrl(), nil)
	if err != nil {
		Logger.Error(err.Error())
	}
	req.Host = c.Request.Hostname
	req.Header.Set("User-Agent", c.Request.UserAgent)

	for k, v := range c.Request.Header {
		req.Header.Set(k, v)
	}
	Logger.Notice("start", c.Title())

	return client.Do(req)
}

func output(resp *http.Response) string {
	var out string
	out += resp.Status + "\n"

	var list []string
	for k, _ := range resp.Header {
		list = append(list, k)
	}

	sort.Strings(list)

	for _, v := range list {
		out += v + ": " + resp.Header[v][0] + "\n"
	}
	return out
}

func Verifier(conf chan *Config, res chan *Result, wg *sync.WaitGroup) {
	var i int
	for {
		select {
		case c := <-conf:
			for i > MaxConn {
				time.Sleep(100 * time.Millisecond)
			}
			i++
			go func(c *Config, res chan *Result, i *int, wg *sync.WaitGroup) {
				res <- c.Verify(i, wg)
			}(c, res, &i, wg)
		}
	}
}

func (c *Config) Verify(i *int, wg *sync.WaitGroup) *Result {
	defer func() {
		Logger.Notice("end verify", c.Title())
		*i--
		wg.Done()
	}()
	Logger.Notice("verify", c.Title())
	result := Result{
		config: c,
	}

	t1 := time.Now()
	resp, err := c.Do()
	result.Duration = time.Since(t1)
	if err != nil {
		result.Error = append(result.Error, err.Error())
		return &result
	}
	defer resp.Body.Close()
	result.rawResp = resp

	if resp.StatusCode == c.Requirement.StatusCode {
		result.Pass = append(result.Pass, fmt.Sprintf("resp code: %d", resp.StatusCode))
	} else {
		result.NotPass = append(result.NotPass, fmt.Sprintf("resp code:%d need:%d", resp.StatusCode, c.Requirement.StatusCode))
	}
	if RawResp {
		var out string
		out += c.Title() + "\n"
		out += output(resp)
		fmt.Println(out)
	}

	for k, v := range c.Requirement.Header {
		val, ok := resp.Header[fixHeader(k)]
		if !ok {
			resp.Header.Set(k, "nil")
			val = []string{"nil"}
		}
		for _, factor := range v {
			h := fmt.Sprint(val)
			factor.sub = h[1 : len(h)-1]
			if h, b, err := factor.Pass(); err == nil {
				if b {
					result.Pass = append(result.Pass, fmt.Sprintf("%s %s %s %s", k, factor.Method, factor.Obj, h))
				} else {
					result.NotPass = append(result.NotPass, fmt.Sprintf("%s %s %s %s not working", k, factor.Method, factor.Obj, h))
				}
			} else {
				result.Error = append(result.Error, err.Error())
				return &result
			}
		}
	}
	return &result
}

func (c *Config) Curl() string {
	out := "curl -s"
	if c.Request.SkipTls {
		out += " -k"
	}
	out += fmt.Sprintf(" -X %s", c.Request.Method)
	out += fmt.Sprintf(" -H Host:%s", c.Request.Hostname)
	out += fmt.Sprintf(" -A '%s'", c.Request.UserAgent)
	for k, v := range c.Request.Header {
		out += fmt.Sprintf(" -H '%s: %s'", k, v)
	}

	out += fmt.Sprintf(" '%s://%s%s'", c.Request.Scheme, c.Request.testIp, c.Request.Uri)

	return out
}

func (r *Result) String() string {
	out := color.Sprintf("@{c}%v %v@{|}\n", r.config.Title(), r.Duration)
	if ShowCurl {
		out += r.config.Curl() + "\n"
	}
	if len(r.Error) != 0 {
		for _, v := range r.Error {
			out += color.Sprintf("@{r}- [✗] %s@{|}\n", v)
		}
	}
	for _, v := range r.Pass {
		out += color.Sprintf("@{g}- [✓] %s@{|}\n", v)
	}
	for _, v := range r.NotPass {
		out += color.Sprintf("@{r}- [✗] %s@{|}\n", v)
	}
	return out
}
