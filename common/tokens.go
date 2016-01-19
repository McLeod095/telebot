package common

import (
	"fmt"
	"time"
)

const tokenSize = 32

type Tokens struct {
	getToken  Token
	postToken Token
	timestamp int64
}

func (t *Tokens) SetGetToken() {
	t.null()
	t.getToken.Gen(tokenSize)
	t.timestamp = time.Now().Unix()
}

func (t *Tokens) GetGetToken() string {
	return t.getToken.Get()
}

func (t *Tokens) GetPostToken() string {
	return t.postToken.Get()
}

func (t *Tokens) SetPostToken() {
	t.getToken.Null()
	t.postToken.Gen(tokenSize)
}

func (t *Tokens) Expired(lifetime int64) (b bool) {
	if time.Now().Unix()-t.timestamp > lifetime {
		b = true
		t.null()
	}
	return
}

func (t *Tokens) End() {
	t.null()
}

func (t *Tokens) null() {
	t.getToken.Null()
	t.postToken.Null()
	t.timestamp = 0
}

func (t *Tokens) String() string {
	return fmt.Sprintf("GetToken: %s PostToken: %s Time: %s", t.getToken, t.postToken, t.timestamp)
}
