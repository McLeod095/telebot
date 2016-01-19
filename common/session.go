package common

import (
	"fmt"
	"github.com/McLeod095/zabbix"
)

type Session struct {
	ChatId     int
	State      int
	Ping       bool
	Messages   MessageCounter
	AuthTokens Tokens
	Api        *zabbix.API
	Media      string
	Alias      string
	AuthToken  Token
}

func (s *Session) GetChatId() int {
	return s.ChatId
}

func (s *Session) GetState() int {
	return s.State
}

func (s *Session) SetState(state int) {
	s.State = state
}

func (s *Session) GetAPI() *zabbix.API {
	return s.Api
}

func (s *Session) SetMedia(m string) {
	s.Media = m
}

func (s *Session) GetMedia() string {
	return s.Media
}

func (s *Session) SetUrl(url string) {
	s.Api = zabbix.NewAPI(url)
}

func (s *Session) SetAuthKey(key, url string) {
	s.SetUrl(url)
	s.Api.Auth = key
}

func (s *Session) String() string {
	return fmt.Sprintf("ChatId:%d Alias:%s State:%d Ping:%t InMessage:%d OutMessage:%d api:%T api:%v",
		s.ChatId, s.Alias, s.State, s.Ping, s.Messages.InCount, s.Messages.OutCount, s.Api, s.Api)
}
