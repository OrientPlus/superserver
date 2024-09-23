package db

import (
	"errors"
	tgapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	psql "superserver/pkg/postgres"
	"github.com/lib/pq"
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
	postgres *psql.Postgres
}

func NewRepo() Repo {
	ri := &repoImpl{}
	ps, err := psql.NewPostgres()
	if err != nil {
		return nil
	}
	ri.postgres = ps

	return ri
}

func (r *repoImpl) AddUser(chat *tgapi.Chat, user *tgapi.User) error {
	if chat == nil {
		return errors.New("value 'chat' is nil")
	}
	if user == nil {
		return errors.New("value 'user' is nil")
	}

	r.postgres.

	return nil
}

func (r *repoImpl) CheckUserAndGroup(chat *tgapi.Chat, user *tgapi.User) bool {

}

func (r *repoImpl) GetUsersFromChat(chat string) ([]tgapi.User, error) {

}

func (r *repoImpl) GetChatParameters(chat string) tgapi.Chat {

}

func (r *repoImpl) GetUserGroups(user tgapi.User) (groups []tgapi.Chat, err error) {

}
