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

// Добавляет юзера в БД. Если такой юзер уже есть, обновляет данные в БД
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

// Обновляет данные юзера в БД
func (r *repoImpl) UpdateUser(user *entity.User) (int64, error) {
	if user == nil {
		return -1, errors.New("value 'user' is nil")
	}

	tx, err := r.pg.BeginTx()
	if err != nil {
		return -1, err
	}

	_, err = r.pg.GetUserByTgID(tx, user.TgID)
	if err != nil {
		rbErr := r.pg.RollbackTx(tx)
		return -1, errors.Join(err, rbErr)
	}

	id, err := r.pg.UpdateUser(tx, *user)
	if err != nil {
		rbErr := r.pg.RollbackTx(tx)
		return -1, errors.Join(err, rbErr)
	}

	err = r.pg.CommitTx(tx)
	return id, err
}

// Возвращает ID юзера из БД
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

// Удаляет юзера из бд
func (r *repoImpl) DeleteUser(user *entity.User) error {
	if user == nil {
		return errors.New("value 'user' is nil")
	}

	tx, err := r.pg.BeginTx()
	if err != nil {
		return err
	}

	err = r.pg.DeleteUser(tx, user.TgID)
	if err != nil {
		rbErr := r.pg.RollbackTx(tx)
		return errors.Join(err, rbErr)
	}

	err = r.pg.CommitTx(tx)

	return err
}

// Добавляет чат в БД и все его производные таблицы
// Если такой чат уже есть, обновляет данные в таблицах
func (r *repoImpl) AddChat(chat *entity.Chat) (int64, error) {
	if chat == nil {
		return -1, errors.New("value 'chat' is nil")
	}

	tx, err := r.pg.BeginTx()
	if err != nil {
		return -1, err
	}

	// Проверяем что такой чат еще не существует
	chatDTO, err := r.pg.GetChat(tx, chat.TgID)
	if err != nil && err != sql.ErrNoRows {
		r.pg.RollbackTx(tx)
		return -1, err
	}
	if err == sql.ErrNoRows {
		r.pg.RollbackTx(tx)
		return r.UpdateChat(chat)
	}

	// Добавляем чат в основную таблицу
	chatDTO.Title = chat.Title
	chatDTO.Type = chat.Type
	chatDTO.TgID = chat.TgID

	chatDTO.LastPesChoice = chat.LastPesChoice
	chatDTO.LastCatChoice = chat.LastCatChoice

	// Проверяем что все юзеры из чата есть в БД
	chatDTO.LastPesID, err = r.pg.AddUser(tx, chat.LastPes)
	if err != nil {
		rbErr := r.pg.RollbackTx(tx)
		return -1, errors.Join(err, rbErr)
	}
	chatDTO.LastCatID, err = r.pg.AddUser(tx, chat.LastCat)
	if err != nil {
		rbErr := r.pg.RollbackTx(tx)
		return -1, errors.Join(err, rbErr)
	}

	chatDTO.OpPerTimeLimiterID, err = r.pg.AddLimiter(tx, psql.LimiterDTO{
		ID:     -1,
		Limit:  float64(chat.OpPerTime.Limit()),
		Burst:  chat.OpPerTime.Burst(),
		Tokens: chat.OpPerTime.Tokens(),
	})
	if err != nil {
		rbErr := r.pg.RollbackTx(tx)
		return -1, errors.Join(err, rbErr)
	}

	chatDTO.LuckyCatLimiterID, err = r.pg.AddLimiter(tx, psql.LimiterDTO{
		ID:     -1,
		Limit:  float64(chat.LuckyCatLimiter.Limit()),
		Burst:  chat.LuckyCatLimiter.Burst(),
		Tokens: chat.LuckyCatLimiter.Tokens(),
	})
	if err != nil {
		rbErr := r.pg.RollbackTx(tx)
		return -1, errors.Join(err, rbErr)
	}

	chatDTO.LuckyPesLimiterID, err = r.pg.AddLimiter(tx, psql.LimiterDTO{
		ID:     -1,
		Limit:  float64(chat.LuckyPesLimiter.Limit()),
		Burst:  chat.LuckyPesLimiter.Burst(),
		Tokens: chat.LuckyPesLimiter.Tokens(),
	})
	if err != nil {
		rbErr := r.pg.RollbackTx(tx)
		return -1, errors.Join(err, rbErr)
	}

	chatId, err := r.pg.AddChat(tx, chatDTO)
	if err != nil {
		rbErr := r.pg.RollbackTx(tx)
		return -1, errors.Join(err, rbErr)
	}

	for _, user := range chat.Members {
		memberTx, err := r.pg.BeginTx()
		if err != nil {
			continue
		}
		userId, err := r.pg.AddUser(memberTx, user)
		if err != nil {
			r.pg.RollbackTx(memberTx)
			continue
		}
		err = r.pg.AddMembers(tx, userId, chatId)
		if err != nil {
			r.pg.RollbackTx(memberTx)
			continue
		}
		r.pg.CommitTx(memberTx)
	}

	for _, event := range chat.Events {
		eventTx, err := r.pg.BeginTx()
		if err != nil {
			continue
		}
		eventDTO := psql.EventDTO{
			ID:         -1,
			CronID:     event.CronID,
			TgID:       event.TgID,
			Title:      event.Title,
			Message:    event.Message,
			TimeConfig: event.TimeConfig,
		}
		eventId, err := r.pg.AddEvent(eventTx, eventDTO)
		if err != nil {
			r.pg.RollbackTx(eventTx)
			continue
		}

		_, err = r.pg.AddEventInChat(eventTx, eventId, chatId)
		if err != nil {
			r.pg.RollbackTx(eventTx)
			continue
		}
		r.pg.CommitTx(eventTx)
	}

	err = tx.Commit()

	return chatId, err
}

func (r *repoImpl) UpdateChat(chat *entity.Chat) (int64, error) {
	if chat == nil {
		return -1, errors.New("value 'chat' is nil")
	}

	tx, err := r.pg.BeginTx()
	if err != nil {
		return -1, err
	}

	// Проверяем что такой чат еще не существует
	chatDTO, err := r.pg.GetChat(tx, chat.TgID)
	if err != nil && err != sql.ErrNoRows {
		r.pg.RollbackTx(tx)
		return -1, err
	}
	if err == sql.ErrNoRows {
		r.pg.RollbackTx(tx)
		return r.AddChat(chat)
	}

	chatDTO.Title = chat.Title
	chatDTO.Type = chat.Type
	chatDTO.TgID = chat.TgID
	chatDTO.LastPesChoice = chat.LastPesChoice
	chatDTO.LastCatChoice = chat.LastCatChoice

	LastPesID, err := r.pg.GetUserIdByTgID(tx, chat.LastPes.TgID)
	if err == nil {
		chatDTO.LastPesID = LastPesID
	}
	LastCatID, err := r.pg.GetUserIdByTgID(tx, chat.LastCat.TgID)
	if err == nil {
		chatDTO.LastCatID = LastCatID
	}

	OpPerTimeLimiterID, err := r.pg.AddLimiter(tx, psql.LimiterDTO{
		ID:     -1,
		Limit:  float64(chat.OpPerTime.Limit()),
		Burst:  chat.OpPerTime.Burst(),
		Tokens: chat.OpPerTime.Tokens(),
	})
	if err == nil {
		chatDTO.OpPerTimeLimiterID = OpPerTimeLimiterID
	}

	LuckyCatLimiterID, err := r.pg.AddLimiter(tx, psql.LimiterDTO{
		ID:     -1,
		Limit:  float64(chat.LuckyCatLimiter.Limit()),
		Burst:  chat.LuckyCatLimiter.Burst(),
		Tokens: chat.LuckyCatLimiter.Tokens(),
	})
	if err == nil {
		chatDTO.LuckyCatLimiterID = LuckyCatLimiterID
	}

	LuckyPesLimiterID, err := r.pg.AddLimiter(tx, psql.LimiterDTO{
		ID:     -1,
		Limit:  float64(chat.LuckyPesLimiter.Limit()),
		Burst:  chat.LuckyPesLimiter.Burst(),
		Tokens: chat.LuckyPesLimiter.Tokens(),
	})
	if err != nil {
		chatDTO.LuckyPesLimiterID = LuckyPesLimiterID
	}

	chatId, err := r.pg.UpdateChat(tx, chatDTO)
	if err != nil {
		rbErr := r.pg.RollbackTx(tx)
		return -1, errors.Join(err, rbErr)
	}

	for _, user := range chat.Members {
		memberTx, err := r.pg.BeginTx()
		if err != nil {
			continue
		}
		userId, err := r.pg.UpdateUser(memberTx, user)
		if err != nil {
			r.pg.RollbackTx(memberTx)
			continue
		}
		err = r.pg.AddMembers(tx, userId, chatId)
		if err != nil {
			r.pg.RollbackTx(memberTx)
			continue
		}
		r.pg.CommitTx(memberTx)
	}

	for _, event := range chat.Events {
		eventTx, err := r.pg.BeginTx()
		if err != nil {
			continue
		}
		eventDTO := psql.EventDTO{
			ID:         -1,
			CronID:     event.CronID,
			TgID:       event.TgID,
			Title:      event.Title,
			Message:    event.Message,
			TimeConfig: event.TimeConfig,
		}
		eventId, err := r.pg.AddEvent(eventTx, eventDTO)
		if err != nil {
			r.pg.RollbackTx(eventTx)
			continue
		}

		_, err = r.pg.AddEventInChat(eventTx, eventId, chatId)
		if err != nil {
			r.pg.RollbackTx(eventTx)
			continue
		}
		r.pg.CommitTx(eventTx)
	}

	err = r.pg.CommitTx(tx)
	return chatId, err
}

func (r *repoImpl) GetChatByTgId(chatId int64) (entity.Chat, error) {
	tx, err := r.pg.BeginTx()
	if err != nil {
		return entity.Chat{}, err
	}

	chatDTO, err := r.pg.GetChat(tx, chatId)
	if err != nil {
		r.pg.RollbackTx(tx)
		return entity.Chat{}, err
	}

	lastCat, err := r.pg.GetUserIdByTgID(tx, chatDTO.LastCatID)
	if err != nil {
		r.pg.RollbackTx(tx)
		return entity.Chat{}, err
	}

	lastPes, err := r.pg.GetUserIdByTgID(tx, chatDTO.LastPesID)
	if err != nil {
		r.pg.RollbackTx(tx)
		return entity.Chat{}, err
	}

	users, err := r.pg.GetChatMembers(tx, chatId)
	if err != nil {
		r.pg.RollbackTx(tx)
		return entity.Chat{}, err
	}

	events, err := r.pg.GetChatEvents(tx, chatId)
	if err != nil {
		r.pg.RollbackTx(tx)
		return entity.Chat{}, err
	}

	opLimiter, err := r.pg.GetLimiterByID(tx, chatDTO.OpPerTimeLimiterID)
	if err != nil {
		r.pg.RollbackTx(tx)
		return entity.Chat{}, err
	}

	luckyCatLimiter, err := r.pg.GetLimiterByID(tx, chatDTO.LuckyCatLimiterID)
	if err != nil {
		r.pg.RollbackTx(tx)
		return entity.Chat{}, err
	}

	luckyPesLimiter, err := r.pg.GetLimiterByID(tx, chatDTO.LuckyPesLimiterID)
	if err != nil {
		r.pg.RollbackTx(tx)
		return entity.Chat{}, err
	}

	return entity.Chat{
		TgID:            chatDTO.TgID,
		Title:           chatDTO.Title,
		Type:            chatDTO.Type,
		LastCat:         lastCat,
		LastPes:         lastPes,
		LastCatChoice:   chatDTO.LastCatChoice,
		LastPesChoice:   chatDTO.LastPesChoice,
		OpPerTime:       opLimiter,
		LuckyCatLimiter: luckyCatLimiter,
		LuckyPesLimiter: luckyPesLimiter,
		Members:         users,
		Events:          events,
	}, err

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
