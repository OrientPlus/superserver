package entity

import (
	tgapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"time"
)

type User struct {
	ID                      int64
	IsBot                   bool
	FirstName               string
	LastName                string
	UserName                string
	LanguageCode            string
	CanJoinGroups           bool
	CanReadAllGroupMessages bool
	SupportsInlineQueries   bool
}

type Chat struct {
	ID                      int64
	Title                   string
	Type                    string
	LastCat                 User
	LastPes                 User
	LastCatChoice           time.Time
	LastPesChoice           time.Time
	LastPressButtonLuckyCat time.Time
	LastPressButtonLuckyPes time.Time
	Members                 []User
}

func GetMessage(upd tgapi.Update) *tgapi.Message {
	var msg *tgapi.Message

	if upd.Message == nil {
		msg = upd.CallbackQuery.Message
	} else {
		msg = upd.Message
	}

	return msg
}
