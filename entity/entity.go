package entity

import (
	tgapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"golang.org/x/time/rate"
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

type ChatEvent struct {
	ID         int
	Title      string
	Message    string
	TimeConfig string
}

type Chat struct {
	ID                      int64
	Title                   string
	Type                    string
	LastCat                 User
	LastPes                 User
	LastCatChoice           time.Time
	LastPesChoice           time.Time
	OpPerTime               *rate.Limiter
	LastPressButtonLuckyCat *rate.Limiter
	LastPressButtonLuckyPes *rate.Limiter
	Members                 []User
	Events                  []ChatEvent
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

func NewUser(user *tgapi.User) User {
	return User{
		ID:                      user.ID,
		IsBot:                   user.IsBot,
		FirstName:               user.FirstName,
		LastName:                user.LastName,
		UserName:                user.UserName,
		LanguageCode:            user.LanguageCode,
		CanJoinGroups:           user.CanJoinGroups,
		CanReadAllGroupMessages: user.CanReadAllGroupMessages,
		SupportsInlineQueries:   user.SupportsInlineQueries,
	}
}

func NewChat(chat *tgapi.Chat) Chat {
	return Chat{
		ID:                      chat.ID,
		Title:                   chat.Title,
		Type:                    chat.Type,
		LastCat:                 User{},
		LastPes:                 User{},
		LastCatChoice:           time.Time{},
		LastPesChoice:           time.Time{},
		OpPerTime:               rate.NewLimiter(1/5, 10),
		LastPressButtonLuckyCat: rate.NewLimiter(1/30, 0),
		LastPressButtonLuckyPes: rate.NewLimiter(1/30, 0),
		Members:                 nil,
		Events:                  nil,
	}
}
