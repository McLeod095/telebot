package notification

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"telebot/common"
	"telebot/models"
)

type ZabbixNotification struct {
	Date    string `json: "date, omitempty"`
	Alias   string `json: "alias"`
	Subject string `json: "subject"`
	Message string `json: "message"`

	EventId   int    `json: "eventid, omitempty"`
	Hostname  string `json: "hostname, omitempty"`
	Ipaddress string `json: "ipaddress, omitempty"`
	Itemvalue string `json: "itemvalue, omitempty"`
	Trigger   struct {
		Id       int    `json:"id, omitempty"`
		Name     string `json:"name, omitempty"`
		Severity string `json:"severity, omitempty"`
		Status   string `json:"status, omitempty"`
		Url      string `json:"url, omitempty"`
	} `json:"trigger, omitempty"`
}

func NotificationHandler(tobot chan<- common.BotMsg) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			body, err := ioutil.ReadAll(r.Body)
			if err != nil {
				log.Println("--", err)
				w.WriteHeader(422)
				return
			}

			err = r.Body.Close()
			if err != nil {
				log.Println("--", err)
				w.WriteHeader(422)
				return
			}

			var notification ZabbixNotification
			err = json.Unmarshal(body, &notification)

			if err != nil {
				log.Println("--", err)
				w.WriteHeader(422)
				return
			}

			sessions, err := models.GetAuthSessionsByAlias(notification.Alias)
			if err != nil {
				log.Println("--", err)
				w.WriteHeader(423)
				return
			}

			var text string = notification.Message

			if len(notification.Trigger.Url) > 0 {
				text += "\n[Описание](" + notification.Trigger.Url + ")"
			}

			for _, s := range *sessions {
				tobot <- common.BotMsg{ChatId: s.ChatId, Text: text, Mode: "Markdown"}
			}

			w.WriteHeader(http.StatusCreated)
		}
	})
}
