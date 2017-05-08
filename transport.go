/* -.-.-.-.-.-.-.-.-.-.-.-.-.-.-.-.-.-.-.-.

* File Name : transport.go

* Purpose :

* Creation Date : 03-26-2016

* Last Modified : Mon May  8 11:59:05 2017

* Created By : Kiyor

_._._._._._._._._._._._._._._._._._._._._.*/

package htest

import (
	"crypto/tls"
	"net"
	"net/http"
	"time"
)

type HTTransport struct {
	Transport http.RoundTripper
	config    *Config
}

func NewHTTransport(c *Config) *HTTransport {
	var keepalive time.Duration
	if c.Request.KeepAlive {
		keepalive = time.Duration(30 * time.Second)
	}

	var tr HTTransport
	tr.Transport = &http.Transport{
		TLSClientConfig: &tls.Config{
			ServerName:         c.Request.Hostname,
			InsecureSkipVerify: c.Request.SkipTls,
		},
		DisableCompression: !c.Request.Compression,
		DisableKeepAlives:  !c.Request.KeepAlive,
		Dial: (&net.Dialer{
			KeepAlive: keepalive,
			Timeout:   c.Request.timeout,
		}).Dial,
		ResponseHeaderTimeout: c.Request.timeout,
	}
	tr.config = c

	return &tr
}

func (c *HTTransport) RoundTrip(req *http.Request) (resp *http.Response, err error) {
	t := c.Transport
	if t == nil {
		t = http.DefaultTransport
	}
	resp, err = t.RoundTrip(req)
	if err != nil {
		return
	}
	switch resp.StatusCode {
	case http.StatusMovedPermanently, http.StatusFound, http.StatusSeeOther, http.StatusTemporaryRedirect:
	}
	return
}
