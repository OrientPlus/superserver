package db

import (
	tgapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/lib/pq"
	_ "github.com/lib/pq"
)

type Repo interface {
	// Добавляет запись о юзере и группе в БД
	AddUser(chat *tgapi.Chat, user *tgapi.User) error

	// Проверяет существование юзера и группы в базе.
	// Если группы или пользователя не существует, вносит соответствующие записи в таблицы
	CheckUserAndGroup(chat *tgapi.Chat, user *tgapi.User) bool

	// Возвращает список всех юзеров состоящих в указанной группе
	GetUsersFromChat(chat string) ([]tgapi.User, error)

	// Возваращает текущие параметры для указанной группы
	GetChatParameters(chat string) tgapi.Chat

	// Возвращает группы в которых состоит указанный юзер
	GetUserGroups(user tgapi.User) (groups []tgapi.Chat, err error)
}

type repoImpl struct {
	postgres pq.Driver
}

func NewRepo() Repo {
	return &repoImpl{}
}

func (r *repoImpl) AddUser(chat *tgapi.Chat, user *tgapi.User) error {

}

func (r *repoImpl) CheckUserAndGroup(chat *tgapi.Chat, user *tgapi.User) bool {

}

func (r *repoImpl) GetUsersFromChat(chat string) ([]tgapi.User, error) {

}

func (r *repoImpl) GetChatParameters(chat string) tgapi.Chat {

}

func (r *repoImpl) GetUserGroups(user tgapi.User) (groups []tgapi.Chat, err error) {

}
