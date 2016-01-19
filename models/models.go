package models

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/spf13/viper"
	"log"
	"telebot/common"
	"time"
)

var db *sql.DB
var dsn string

const rTime = 30

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}

func init() {
	viper.SetConfigName("telebot")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("$HOME/.telebot")

	err := viper.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("Fatal error config file: %s \n", err))
	}

	dsn = fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8",
		viper.GetString("mysql.telebot.user"),
		viper.GetString("mysql.telebot.password"),
		viper.GetString("mysql.telebot.host"),
		viper.GetString("mysql.telebot.port"),
		viper.GetString("mysql.telebot.dbname"))

	db, err = sql.Open("mysql", dsn)
	checkErr(err)

	err = db.Ping()
	checkErr(err)

	go func() {
		tick := time.Tick(rTime * time.Second)
		for range tick {
			if err := db.Ping(); err != nil {
				log.Println("Connection to DB lost (try reconnect after", rTime, "seconds)")
			}
		}
	}()
}

func GetSession(id int) (*common.Session, error) {
	const query = "SELECT chatid, state, ping, media, alias, token FROM sessions WHERE chatid=?"

	var ret common.Session
	var token string
	err := db.QueryRow(query, id).Scan(&ret.ChatId, &ret.State, &ret.Ping, &ret.Media, &ret.Alias, &token)
	ret.AuthToken.Set(token)

	return &ret, err
}

func GetAuthSession(id int) (*common.Session, error) {
	const query = "SELECT chatid, state, ping, media, alias, token FROM sessions WHERE chatid=? AND state=?"

	var ret common.Session
	var token string
	err := db.QueryRow(query, id, 2).Scan(&ret.ChatId, &ret.State, &ret.Ping, &ret.Media, &ret.Alias, &token)
	ret.AuthToken.Set(token)

	return &ret, err
}

func GetAuthSessions() (*[]common.Session, error) {
	const query = "SELECT chatid, state, ping, media, alias FROM sessions WHERE state=?"

	var ret []common.Session

	rows, err := db.Query(query, 2)
	if err == nil {
		for rows.Next() {
			var t common.Session
			err = rows.Scan(&t.ChatId, &t.State, &t.Ping, &t.Media, &t.Alias)
			if err != nil {
				return nil, err
			}
			ret = append(ret, t)
		}
	}

	return &ret, err
}

func GetPingSessions() (*[]common.Session, error) {
	const query = "SELECT chatid, state, ping, media, alias FROM sessions WHERE state=? AND ping=?"

	var ret []common.Session

	rows, err := db.Query(query, 2, 1)
	if err == nil {
		for rows.Next() {
			var t common.Session
			if err := rows.Scan(&t.ChatId, &t.State, &t.Ping, &t.Media, &t.Alias); err != nil {

			}
			ret = append(ret, t)
		}
	}

	return &ret, err
}

func GetSessionsByAlias(alias string) (*[]common.Session, error) {
	const query = "SELECT chatid, state, ping, media, alias FROM sessions WHERE alias=?"

	var ret []common.Session

	rows, err := db.Query(query, alias)
	if err == nil {
		for rows.Next() {
			var t common.Session
			err = rows.Scan(&t.ChatId, &t.State, &t.Ping, &t.Media, &t.Alias)
			if err != nil {
				return nil, err
			}
			ret = append(ret, t)
		}
	}

	return &ret, err
}

func GetAuthSessionsByAlias(alias string) (*[]common.Session, error) {
	const query = "SELECT chatid, state, ping, media, alias FROM sessions WHERE alias=? AND state=?"

	var ret []common.Session

	rows, err := db.Query(query, alias, 2)
	if err == nil {
		for rows.Next() {
			var t common.Session
			if err := rows.Scan(&t.ChatId, &t.State, &t.Ping, &t.Media, &t.Alias); err != nil {

			}
			ret = append(ret, t)
		}
	}

	return &ret, err
}

func AddSession(s *common.Session) (err error) {
	const query = "INSERT INTO sessions (chatid, state, ping, media, alias, token) VALUES (?, ?, ?, ?, ?, ?)"

	stmt, err := db.Prepare(query)
	if err != nil {
		return
	}

	_, err = stmt.Exec(
		s.ChatId,
		s.State,
		s.Ping,
		s.Media,
		s.Alias,
		string(s.AuthToken),
	)
	return
}

func UpdateSession(s *common.Session) (err error) {
	const query = "UPDATE sessions set state=?, ping=?, media=?, alias=?, token=? WHERE chatid=?"

	stmt, err := db.Prepare(query)
	if err != nil {
		return
	}

	_, err = stmt.Exec(
		s.State,
		s.Ping,
		s.Media,
		s.Alias,
		string(s.AuthToken),
		s.ChatId,
	)
	return
}
