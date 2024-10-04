package entity

import (
	tgapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"golang.org/x/time/rate"
	"time"
)

type User struct {
	TgID                    int64
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
	TgID            int64
	Title           string
	Type            string
	LastCat         User
	LastPes         User
	LastCatChoice   time.Time
	LastPesChoice   time.Time
	OpPerTime       *rate.Limiter
	LuckyCatLimiter *rate.Limiter
	LuckyPesLimiter *rate.Limiter
	Members         []User
	Events          []ChatEvent
}

type ChatEvent struct {
	CronID     int64
	TgID       int64
	Title      string
	Message    string
	TimeConfig string
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
		TgID:                    user.ID,
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
		TgID:            chat.ID,
		Title:           chat.Title,
		Type:            chat.Type,
		LastCat:         User{},
		LastPes:         User{},
		LastCatChoice:   time.Time{},
		LastPesChoice:   time.Time{},
		OpPerTime:       rate.NewLimiter(1/5, 10),
		LuckyCatLimiter: rate.NewLimiter(1/30, 0),
		LuckyPesLimiter: rate.NewLimiter(1/30, 0),
		Members:         nil,
		Events:          nil,
	}
}
