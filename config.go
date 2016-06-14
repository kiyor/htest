/* -.-.-.-.-.-.-.-.-.-.-.-.-.-.-.-.-.-.-.-.

* File Name : config.go

* Purpose :

* Creation Date : 03-25-2016

* Last Modified : Fri Jun  3 17:24:07 2016

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
	"strings"
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
	templateMux = new(sync.Mutex)
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
	timeout     time.Duration
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
func saveTemplate(c *Config, prefix string) {
	if len(c.Hash) == 0 {
		Logger.Warning("Template will only load the config with hash", c.Title())
	} else {
		if _, ok := TemplateMap[prefix+c.Hash]; ok {
			Logger.Warning("Template already loaded, ignore", c.Title())
		} else {
			templateMux.Lock()
			TemplateMap[c.Hash] = c
			templateMux.Unlock()
		}
	}
}
func loadTemplate(file string) {
	var config Config
	var configs []*Config

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
		saveTemplate(&config, "")
	}
	// config is a list of config
	if err2 == nil {
		for _, c := range configs {
			saveTemplate(c, "")
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

func cleanConfig(c *Config, ips ...string) []*Config {
	var res []*Config

	if len(ips) != 0 {
		for _, ip := range ips {
			newc := *c
			newc.Request.testIp = ip
			res = append(res, &newc)
		}
		ips = []string{}
		var res2 []*Config
		for _, v := range res {
			res2 = append(res2, cleanConfig(v)...)
		}
		return res2
	}

	if c.Request.Scheme == "both" {
		c1 := *c
		c2 := *c
		c1.Request.Scheme = "http"
		c2.Request.Scheme = "https"
		var res2 []*Config
		res2 = append(res2, cleanConfig(&c1)...)
		res2 = append(res2, cleanConfig(&c2)...)
		return res2
	}

	if strings.Contains(c.Request.Hostname, " ") {
		str := c.Request.Hostname
		for strings.Contains(str, "  ") {
			str = strings.Replace(str, "  ", " ", -1)
		}
		var res2 []*Config
		for _, v := range strings.Split(str, " ") {
			newc := *c
			newc.Request.Hostname = v
			res2 = append(res2, cleanConfig(&newc)...)
		}
		return res2
	}
	if strings.Contains(c.Request.Uri, " ") {
		str := c.Request.Uri
		for strings.Contains(str, "  ") {
			str = strings.Replace(str, "  ", " ", -1)
		}
		var res2 []*Config
		for _, v := range strings.Split(str, " ") {
			newc := *c
			newc.Request.Uri = v
			res2 = append(res2, cleanConfig(&newc)...)
		}
		return res2
	}
	if len(res) == 0 {
		return []*Config{c}
	}
	return res
}

func doCheck(file string, configChan chan *Config, results chan *Result, wg *sync.WaitGroup, ips ...string) {
	var config Config
	var configs []*Config

	sendQueue := func(c *Config) {
		wg.Add(1)
		configChan <- c
	}

	check := func(c *Config) {
		if len(c.Hash) != 0 && len(c.Request.Hostname) == 0 {
			saveTemplate(c, file+":")
			return
		}
		cs := cleanConfig(c, ips...)
		// 		Logger.Error(len(cs))
		for _, v := range cs {
			sendQueue(v)
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
		Logger.Error(file, err1.Error(), "\n", err2.Error())
		return
	}

	configInherit := func(config Config) *Config {
		requestHeader := config.Request.Header
		newRequestHeader := make(map[string]string)
		for _, in := range config.Request.Include {
			localIn := file + ":" + in
			if template, ok := TemplateMap[localIn]; ok {
				for k, v := range template.Request.Header {
					newRequestHeader[k] = v
				}
			} else {
				if template, ok := TemplateMap[in]; ok {
					for k, v := range template.Request.Header {
						newRequestHeader[k] = v
					}
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
			localIn := file + ":" + in
			if template, ok := TemplateMap[localIn]; ok {
				for k, v := range template.Requirement.Header {
					for _, v1 := range v {
						newRequirementHeader[k] = append(newRequirementHeader[k], v1)
					}
				}
			} else {
				if template, ok := TemplateMap[in]; ok {
					for k, v := range template.Requirement.Header {
						for _, v1 := range v {
							newRequirementHeader[k] = append(newRequirementHeader[k], v1)
						}
					}
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
	var err error
	if len(OverTimeout) > 0 {
		c.Request.Timeout = OverTimeout
	}
	c.Request.timeout, err = time.ParseDuration(c.Request.Timeout)
	if err != nil {
		c.Request.timeout = time.Duration(10 * time.Second)
	}

	client := &http.Client{
		Transport: NewHTTransport(c),
		Timeout:   c.Request.timeout,
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

	// 	return client.Do(req)
	return client.Transport.RoundTrip(req)
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
	if c.Request.KeepAlive {
		out += " --keepalive-time 30"
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
	list1, list2 := []string{}, []string{}
	for _, v := range r.Pass {
		list1 = append(list1, v)
	}
	for _, v := range r.NotPass {
		list2 = append(list2, v)
	}
	sort.Strings(list1)
	sort.Strings(list2)
	// 	var i int
	for _, v := range list1 {
		// 		i++
		// 		if v[:1] != " " {
		// 			out += color.Sprintf("@{g}- [✓] %2d. %s@{|}\n", i, v)
		// 		} else {
		out += color.Sprintf("@{g}- [✓] %s@{|}\n", v)
		// 		}
	}
	for _, v := range list2 {
		// 		i++
		// 		if v[:1] != " " {
		// 			out += color.Sprintf("@{r}- [✗] %2d. %s@{|}\n", i, v)
		// 		} else {
		out += color.Sprintf("@{r}- [✗] %s@{|}\n", v)
		// 		}
	}
	return out
}
