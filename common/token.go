package common

import (
	"fmt"
	"math/rand"
	"time"
)

const (
	letterBytes   = ("abcdefghijklmnopqrstuvwxyz0123456789")
	letterIdxBits = 6                    // 6 bits to represent a letter index
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
)

var src = rand.NewSource(time.Now().UnixNano())

type Token string

func (t *Token) Gen(n int) string {
	b := make([]byte, n)
	for i, cache, remain := n-1, src.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = src.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
			b[i] = letterBytes[idx]
			i--
		}
		cache >>= letterIdxBits
		remain--
	}

	(*t) = Token(b)
	return string(*t)
}

func (t Token) Get() string {
	return string(t)
}

func (t *Token) Set(s string) {
	(*t) = Token(s)
}

func (t *Token) Null() {
	(*t) = Token("")
}

func (t Token) String() string {
	return fmt.Sprintf("%s", string(t))
}
