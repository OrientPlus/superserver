package db

import (
	"errors"
	"superserver/entity"
	psql "superserver/pkg/postgres"
)

type Repo interface {
	// Добавляет запись о юзере и группе в БД
	AddUser(chat *entity.Chat, user *entity.User) error

	// Проверяет существование юзера и группы в базе.
	// Если группы или пользователя не существует, вносит соответствующие записи в таблицы
	CheckUserAndGroup(chat *entity.Chat, user *entity.User) bool

	// Возвращает список всех юзеров состоящих в указанной группе
	GetUsersFromChat(chat string) ([]entity.User, error)

	// Возваращает текущие параметры для указанной группы
	GetChatParameters(chat string) entity.Chat

	// Возвращает группы в которых состоит указанный юзер
	GetUserGroups(user entity.User) (groups []entity.Chat, err error)
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

func (r *repoImpl) AddUser(chat *entity.Chat, user *entity.User) error {
	if chat == nil {
		return errors.New("value 'chat' is nil")
	}
	if user == nil {
		return errors.New("value 'user' is nil")
	}

	tx, err := r.postgres.BeginTx()
	if err != nil {
		return err
	}

	// Узнаем нет ли уже такой группы
	// TODO корректная обработка события когда записи нет в БД, но возвращается ошибка
	group, err := r.postgres.GetChat(tx, chat.TgID)
	if err != nil {
		r.postgres.RollbackTx(tx)
		return err
	}

	// Если нет - добавляем группу

	// Если есть обновляем данные группы

	// Добавляем юзера
	r.postgres.AddUser(tx, *user)

	return nil
}

func (r *repoImpl) CheckUserAndGroup(chat *entity.Chat, user *entity.User) bool {

}

func (r *repoImpl) GetUsersFromChat(chat string) ([]entity.User, error) {

}

func (r *repoImpl) GetChatParameters(chat string) entity.Chat {

}

func (r *repoImpl) GetUserGroups(user entity.User) (groups []entity.Chat, err error) {

}
