/* -.-.-.-.-.-.-.-.-.-.-.-.-.-.-.-.-.-.-.-.

* File Name : token.go

* Purpose :

* Creation Date : 03-09-2018

* Last Modified : Fri Mar  9 19:19:55 2018

* Created By : Kiyor

_._._._._._._._._._._._._._._._._._._._._.*/

package token

import (
	"crypto/md5"
	"encoding/hex"
	"net/http"
	"strings"
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
	u = strings.Replace(u, r.URL.Host, r.Host, -1)
	var s string
	if len(r.URL.Query()) != 0 {
		s = "&"
	}
	r.URL.RawQuery += s + "h=" + md(token.Key+u)
}
