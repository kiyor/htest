/* -.-.-.-.-.-.-.-.-.-.-.-.-.-.-.-.-.-.-.-.

* File Name : token.go

* Purpose :

* Creation Date : 03-09-2018

* Last Modified : Fri Mar  9 19:19:55 2018

* Created By : Kiyor

_._._._._._._._._._._._._._._._._._._._._.*/

package token

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

type A1 struct {
	Key string
}

func (A1) sha256(data string) string {
	h := sha256.New()
	h.Write([]byte(data))
	return fmt.Sprintf("%x", h.Sum(nil))
}

func (a1 *A1) encryptCBC(plaintext []byte, keystr string) (crypted []byte, err error) {
	key, err := hex.DecodeString(keystr)
	if err != nil {
		return nil, err
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	iv := []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}

	cbc := cipher.NewCBCEncrypter(block, iv)
	content := []byte(plaintext)
	content = PKCS5Padding(content, block.BlockSize())
	crypted = make([]byte, len(content))
	cbc.CryptBlocks(crypted, content)

	return
}

func PKCS5Padding(ciphertext []byte, blockSize int) []byte {
	padding := blockSize - len(ciphertext)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(ciphertext, padtext...)
}

func (a *A1) Apply(r *http.Request) {
	expire := time.Now().Add(1 * time.Hour).Unix()
	random := "test"
	text := fmt.Sprintf("%s%d%s", r.URL.RequestURI(), expire, random)
	shakey := a.sha256(a.Key + random)
	ciphertext, err := a.encryptCBC([]byte(text), shakey)
	if err != nil {
		panic(err)
	}
	cleartext := base64.StdEncoding.EncodeToString(ciphertext)
	var s string
	if len(r.URL.Query()) != 0 {
		s = "&"
	}
	cleartext = url.QueryEscape(cleartext)
	r.URL.RawQuery += s + "accessKey=" + fmt.Sprint(expire) + "_" + random + "_" + cleartext
}
