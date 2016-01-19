package main

import (
	"encoding/json"
	"fmt"
	"github.com/McLeod095/zabbix"
	"github.com/Syfaro/telegram-bot-api"
	"github.com/spf13/viper"
	"github.com/streadway/amqp"
	"log"
	"net/http"
	"net/mail"
	"net/smtp"
	"strings"
	"telebot/common"
	"telebot/models"
	"telebot/notification"
	"time"
)

type Message struct {
	User    string `json:user`
	Subject string `json:subject`
	Body    string `json:body`
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
		panic(fmt.Sprintf("%s: %s", msg, err))
	}
}

var State map[string]int
var StringState map[int]string
var Signal map[string]int
var States [][]int
var medias zabbix.Medias
var token int

func SendEmail(email, token string) (err error) {
	client, err := smtp.Dial(viper.GetString("smtp.host") + ":" + viper.GetString("smtp.port"))
	if err != nil {
		log.Println(err)
		return
	}

	err = client.Mail(viper.GetString("smtp.mailfrom"))
	if err != nil {
		log.Println(err)
		return
	}

	err = client.Rcpt(email)
	if err != nil {
		log.Println(err)
		return
	}

	wc, err := client.Data()
	if err != nil {
		log.Println(err)
		return
	}

	fmt.Fprintf(wc, "Subject: Token for zabbix bot\n\n")
	fmt.Fprintf(wc, "Generated token %s\n", token)
	fmt.Fprintf(wc, "Use in chat\n")
	fmt.Fprintf(wc, "/token %s\n", token)

	err = wc.Close()
	if err != nil {
		log.Println(err)
		return
	}

	err = client.Quit()
	if err != nil {
		log.Println(err)
		return
	}
	return
}

func main() {
	viper.SetConfigName("telebot")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("$HOME/.telebot")

	err := viper.ReadInConfig()

	if err != nil {
		panic(fmt.Errorf("Fatal error config file: %s \n", err))
	}

	api := zabbix.NewAPI(viper.GetString("zabbix.proto") + "://" +
		viper.GetString("zabbix.host") + ":" +
		viper.GetString("zabbix.port") + "/api_jsonrpc.php")
	api.Login(viper.GetString("zabbix.user"), viper.GetString("zabbix.password"))

	StringState = map[int]string{0: "NotAuth", 1: "Token", 2: "Auth"}
	State = map[string]int{"NotAuth": 0, "Token": 1, "Auth": 2}
	Signal = map[string]int{"Login": 0, "Token": 1, "Logout": 2}
	States = [][]int{{1, 0, 0}, {0, 2, 0}, {2, 2, 0}}

	bot, err := tgbotapi.NewBotAPI(viper.GetString("telegram.botkey"))
	log.Println(viper.GetString("telegram.botkey"))
	if err != nil {
		log.Panic(err)
	}

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)

	//	from_bot := make(chan common.BotMsg, 0)
	to_bot := make(chan common.BotMsg, 0)
	quit := make(chan int, 0)

	go func() {
		for {
			select {
			case tobot := <-to_bot:
				msg := tgbotapi.NewMessage(tobot.ChatId, tobot.Text)
				if len(tobot.Mode) > 0 {
					msg.ParseMode = tobot.Mode
				}
				log.Println("->", msg.ChatID, msg.Text)
				bot.Send(msg)
			}
		}
	}()

	go func() {
		for {
			select {
			case update := <-updates:
				Message := update.Message
				log.Println("<-", Message.Chat.ID, Message.Text)

				if Message.IsCommand() {
					switch Message.Command() {
					case "/login":
						go func() {
							users, err := api.UsersGetEnabled(zabbix.Params{})
							if err != nil {
								log.Println(err)
								return
							}

							var userids []string
							for _, user := range users {
								userids = append(userids, user.UserId)
							}

							medias, err := api.MediasByUserIds(userids)

							if err != nil {
								log.Println(err)
								return
							}

							for _, media := range medias {
								if Message.CommandArguments() == media.Sendto && (media.MediatypeId == "1" || media.MediatypeId == "5") && media.Active == 0 {
									log.Println("dd", Message.CommandArguments(), media.Sendto)
									var session common.Session
									session.ChatId = Message.Chat.ID
									session.Media = media.Sendto
									var alias string
									for _, u := range users {
										if u.UserId == media.UserId {
											alias = u.Alias
											break
										}
									}
									//session.Alias = Message.Chat.UserName
									session.Alias = alias
									log.Println("--", session.ChatId, "Change state from", StringState[session.State], "to", StringState[States[session.State][Signal["login"]]])
									session.State = States[session.State][Signal["Login"]]
									if session.State == Signal["Token"] {
										session.AuthToken.Gen(32)
										log.Println("--", "Generated token", session.AuthToken.Get())
										if _, err := mail.ParseAddress(media.Sendto); err == nil {
											if err := SendEmail(media.Sendto, session.AuthToken.Get()); err == nil {
												var text string
												if Message.Chat.FirstName != "" && Message.Chat.LastName != "" {
													text = Message.Chat.FirstName + " " + Message.Chat.LastName + " ключ выслан"
												}
												to_bot <- common.BotMsg{Text: text, ChatId: Message.Chat.ID}
											}
										}
									}
									if err := models.AddSession(&session); err != nil {
										log.Println(err)
									}

									break
								}
							}
						}()
					case "/ping":
						if s, err := models.GetSession(Message.Chat.ID); err == nil {
							s.Ping = !s.Ping
							if err := models.UpdateSession(s); err != nil {
								log.Println(err)
							}
							var text string
							if s.Ping {
								text = "Ping On"
							} else {
								text = "Ping Off"
							}
							to_bot <- common.BotMsg{Text: text, ChatId: s.ChatId}
						}
					case "/token":
						session, err := models.GetSession(Message.Chat.ID)
						if err == nil {
							log.Println("dd", session)
							Token := Message.CommandArguments()
							if Token == session.AuthToken.Get() {
								session.State = States[session.State][Signal["Token"]]
								session.AuthToken.Null()
								if err := models.UpdateSession(session); err != nil {
									log.Println(err)
								} else {
									to_bot <- common.BotMsg{Text: StringState[session.State], ChatId: session.ChatId}
								}
							}
						} else {
							log.Println("dd", err)
						}
					case "/who":
						if _, err := models.GetAuthSession(Message.Chat.ID); err == nil {
							if users, err := models.GetAuthSessions(); err == nil {
								var text []string
								for _, u := range *users {
									text = append(text, u.Alias)
								}
								to_bot <- common.BotMsg{ChatId: Message.Chat.ID, Text: strings.Join(text, "\n")}
							}
						}
					case "/logout":
					case "/to":
						_, err := models.GetAuthSession(Message.Chat.ID)
						if err != nil {
							log.Println(err)
							break
						}

						if len(Message.CommandArguments()) == 0 {
							log.Println("--", "Command wtihout arguments")
							break
						}

						alias := strings.Split(Message.CommandArguments(), " ")[0]
						message := strings.Split(Message.CommandArguments(), " ")[1:]

						if len(message) == 0 {
							log.Println("--", "Message is null")
						}

						sessions, err := models.GetSessionsByAlias(alias)
						if err != nil {
							log.Println(err)
							break
						}

						from, _ := models.GetAuthSession(Message.Chat.ID)
						for _, session := range *sessions {
							to_bot <- common.BotMsg{ChatId: session.ChatId, Text: fmt.Sprintf("From: %s\n%s\n", from.Alias, strings.Join(message, " "))}
						}

					case "/status":
						text := "Not Exists"
						if session, err := models.GetSession(Message.Chat.ID); err == nil {
							text = StringState[session.State]
						}
						to_bot <- common.BotMsg{Text: text, ChatId: Message.Chat.ID}
					default:
						log.Println("--", "Unknow command", Message.Command(), Message.CommandArguments())
					}
				} else {
					if s, err := models.GetSession(Message.Chat.ID); err == nil && s.State == State["Auth"] {
						to_bot <- common.BotMsg{Text: Message.Text, ChatId: Message.Chat.ID}
					} else {
						log.Println("--", "Message", Message.Chat.ID, Message.Text)
					}
				}
			}
		}
	}()

	go func() {
		tick := time.Tick(1 * time.Minute)
		for range tick {
			sessions, err := models.GetPingSessions()
			if err == nil {
				for _, s := range *sessions {
					to_bot <- common.BotMsg{ChatId: s.ChatId, Text: "Ping"}
				}
			} else {
				log.Println("dd", err)
			}
		}
	}()

	conn, err := amqp.Dial("amqp://" +
		viper.GetString("rabbitmq.user") + ":" +
		viper.GetString("rabbitmq.password") + "@" +
		viper.GetString("rabbitmq.host") + ":" +
		viper.GetString("rabbitmq.port") + "/" +
		viper.GetString("rabbitmq.path"))

	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	q, err := ch.QueueDeclare(
		viper.GetString("rabbitmq.queue"), // name
		false, // durable
		false, // delete when usused
		false, // exclusive
		false, // no-wait
		nil,   // arguments
	)
	failOnError(err, "Failed to declare a queue")

	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		true,   // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	failOnError(err, "Failed to register a consumer")

	go func() {
		for d := range msgs {
			var m Message
			err := json.Unmarshal(d.Body, &m)
			log.Printf("Received a message: %s", d.Body)

			if err == nil {
				sessions, err := models.GetAuthSessionsByAlias(m.User)
				if err == nil {
					for _, s := range *sessions {
						to_bot <- common.BotMsg{ChatId: s.ChatId, Text: fmt.Sprintf("*%s*\n%s", m.Subject, m.Body), Mode: "Markdown"}
					}
				} else {
					log.Println("dd", err)
				}
			} else {
				log.Println("unMarshal error")
			}
		}
	}()

	http.Handle(viper.GetString("http.locations.notification"), notification.NotificationHandler(to_bot))
	http.ListenAndServe(viper.GetString("http.ip")+":"+viper.GetString("http.port"), nil)
	<-quit
}
