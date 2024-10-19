package entity

import (
	"container/list"
	"time"

	tgapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"golang.org/x/time/rate"
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
	LastCat         *User
	LastPes         *User
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
		LastCat:         nil,
		LastPes:         nil,
		LastCatChoice:   time.Time{},
		LastPesChoice:   time.Time{},
		OpPerTime:       rate.NewLimiter(1/5, 10),
		LuckyCatLimiter: rate.NewLimiter(1/30, 0),
		LuckyPesLimiter: rate.NewLimiter(1/30, 0),
		Members:         nil,
		Events:          nil,
	}
}

type ChatsQueue struct {
	items     *list.List
	Length    int
	MaxLength int
}

// NewQueue создает и возвращает новую очередь
func NewQueue() *ChatsQueue {
	return &ChatsQueue{
		items:     list.New(),
		Length:    0,
		MaxLength: 5,
	}
}

// Push добавляет элемент в конец очереди
func (q *ChatsQueue) Push(value Chat) {
	if q.Length == q.MaxLength {
		element := q.items.Front()
		q.items.Remove(element)
	} else {
		q.Length++
	}

	q.items.PushBack(value)
}

// Pop удаляет элемент из начала очереди и возвращает его
func (q *ChatsQueue) Pop() (Chat, bool) {
	if q.Length == 0 {
		return Chat{}, false
	}

	element := q.items.Front()
	value := element.Value
	q.items.Remove(element)
	q.Length--
	chat, _ := value.(Chat)
	return chat, true
}

// Exist проверяет, существует ли элемент в очереди
func (q *ChatsQueue) Exist(TgId int64) (Chat, bool) {
	for element := q.items.Front(); element != nil; element = element.Next() {
		cur_id := element.Value.(Chat).TgID
		if cur_id == TgId {
			return Chat{}, true
		}
	}
	return Chat{}, false
}
