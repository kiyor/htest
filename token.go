/* -.-.-.-.-.-.-.-.-.-.-.-.-.-.-.-.-.-.-.-.

* File Name : token.go

* Purpose :

* Creation Date : 03-09-2018

* Last Modified : Fri Mar  9 17:53:54 2018

* Created By : Kiyor

_._._._._._._._._._._._._._._._._._._._._.*/

package htest

import (
	"crypto/md5"
	"encoding/hex"
	"net/http"
)

type LL1 struct {
	Key string
}

func md(text string) string {
	hasher := md5.New()
	hasher.Write([]byte(text))
	return hex.EncodeToString(hasher.Sum(nil))
}

func (token *LL1) Apply(r *http.Request) {
	u := r.URL.String()
	var s string
	if len(r.URL.Query()) != 0 {
		s = "&"
	}
	r.URL.RawQuery += s + "h=" + md(token.Key+u)
}
