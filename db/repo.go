package db

import (
	"database/sql"
	"errors"

	"superserver/entity"
	psql "superserver/pkg/postgres"
)

type Repo interface {
	// Добавляет запись о юзере и группе в БД
	AddUser(user *entity.User) (int64, error)

	UpdateUser(user *entity.User) (int64, error)

	GetUserId(user *entity.User) (int64, error)

	DeleteUser(user *entity.User) error

	// Добавляем чат в БД
	AddChat(chat *entity.Chat) (int64, error)

	// Обновляет данные о чате в БД
	UpdateChat(chat *entity.Chat) (int64, error)

	GetChatId(chat *entity.Chat) (int64, error)

	DeleteChat(chat *entity.Chat) error

	AddUserInChat(userId, chatID int64) (int64, error)

	DeleteUserFromChat(userId, chatId int64) error

	AddEvent(event entity.ChatEvent) (int64, error)

	GetEvent(event entity.ChatEvent) (int64, error)

	UpdateEvent(event entity.ChatEvent) (int64, error)

	DeleteEvent(event entity.ChatEvent) error

	// @eventId - db uniq id
	// @chatId  - db uniq id
	AddEventInChat(eventId, chatId int64) (int64, error)

	// @eventId - db uniq id
	// @chatId  - db uniq id
	DeleteEventFromChat(eventId, chatId int64) error

	// @cronEventId - cron event uniq id
	// @chatId  	- tg uniq id
	DeleteEventFromChatByExternalId(cronEventId, chatId int64) error

	// Получить список групп пользователя
	GetUserGroups(user *entity.User) ([]entity.Chat, error)

	// Возвращает список всех юзеров состоящих в указанной группе
	GetUsersFromChat(chat string) ([]entity.User, error)

	// Возваращает текущие параметры для указанной группы
	GetChatParameters(chat string) (entity.Chat, int64)
}

type repoImpl struct {
	pg *psql.Postgres
}

func (r *repoImpl) AddUser(user *entity.User) (int64, error) {
	if user == nil {
		return -1, errors.New("value 'user' is nil")
	}

	tx, err := r.pg.BeginTx()
	if err != nil {
		return -1, err
	}

	// Проверяем что юзера нет в БД
	_, err = r.pg.GetUserByTgID(tx, user.TgID)
	if err != nil && err != sql.ErrNoRows {
		rbErr := r.pg.RollbackTx(tx)
		return -1, errors.Join(err, rbErr)
	}

	if err == nil {
		return r.UpdateUser(user)
	}

	// Добавляем юзера
	userId, err := r.pg.AddUser(tx, *user)
	if err != nil {
		rbErr := r.pg.RollbackTx(tx)
		return -1, errors.Join(err, rbErr)
	}

	err = r.pg.CommitTx(tx)
	return userId, err
}

func (r *repoImpl) UpdateUser(user *entity.User) (int64, error) {
	if user == nil {
		return -1, errors.New("value 'user' is nil")
	}

	tx, err := r.pg.BeginTx()
	if err != nil {
		return -1, err
	}

	id, err := r.pg.UpdateUser(tx, *user)
	if err != nil {
		rbErr := r.pg.RollbackTx(tx)
		return -1, errors.Join(err, rbErr)
	}

	err = r.pg.CommitTx(tx)
	return id, err
}

func (r *repoImpl) GetUserId(user *entity.User) (int64, error) {
	tx, err := r.pg.BeginTx()
	if err != nil {
		return -1, err
	}

	userID, err := r.pg.GetUserIdByTgID(tx, user.TgID)
	if err != nil {
		rbErr := r.pg.RollbackTx(tx)
		return -1, errors.Join(err, rbErr)
	}

	err = r.pg.CommitTx(tx)
	return userID, err
}

func (r *repoImpl) DeleteUser(user *entity.User) error {
	if user == nil {
		return errors.New("value 'user' is nil")
	}

	tx, err := r.pg.BeginTx()
	if err != nil {
		return err
	}

	err = r.pg.DeleteUser(tx, user.TgID)

	return err
}

func (r *repoImpl) AddChat(chat *entity.Chat) (int64, error) {
	//TODO implement me
	panic("implement me")
}

func (r *repoImpl) GetChatId(chat *entity.Chat) (int64, error) {
	//TODO implement me
	panic("implement me")
}

func (r *repoImpl) DeleteChat(chat *entity.Chat) error {
	//TODO implement me
	panic("implement me")
}

func (r *repoImpl) AddUserInChat(userId, chatID int64) (int64, error) {
	//TODO implement me
	panic("implement me")
}

func (r *repoImpl) DeleteUserFromChat(userId, chatId int64) error {
	//TODO implement me
	panic("implement me")
}

func (r *repoImpl) AddEvent(event entity.ChatEvent) (int64, error) {
	//TODO implement me
	panic("implement me")
}

func (r *repoImpl) GetEvent(event entity.ChatEvent) (int64, error) {
	//TODO implement me
	panic("implement me")
}

func (r *repoImpl) UpdateEvent(event entity.ChatEvent) (int64, error) {
	//TODO implement me
	panic("implement me")
}

func (r *repoImpl) DeleteEvent(event entity.ChatEvent) error {
	//TODO implement me
	panic("implement me")
}

func (r *repoImpl) AddEventInChat(eventId, chatId int64) (int64, error) {
	//TODO implement me
	panic("implement me")
}

func (r *repoImpl) DeleteEventFromChat(eventId, chatId int64) error {
	//TODO implement me
	panic("implement me")
}

func (r *repoImpl) GetUserGroups(user *entity.User) ([]entity.Chat, error) {
	//TODO implement me
	panic("implement me")
}

func NewRepo() Repo {
	ri := &repoImpl{}
	ps, err := psql.NewPostgres()
	if err != nil {
		return nil
	}
	ri.pg = ps

	return ri
}
