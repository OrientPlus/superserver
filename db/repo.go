package db

import (
	"database/sql"
	"errors"

	"golang.org/x/time/rate"
	"superserver/entity"
	psql "superserver/pkg/postgres"
)

type Repo interface {
	BeginTx() (*sql.Tx, error)
	CommitTx(tx *sql.Tx) error
	RollbackTx(tx *sql.Tx) error

	// Добавляет запись о юзере и группе в БД
	AddUser(tx *sql.Tx, user *entity.User) (int64, error)

	UpdateUser(tx *sql.Tx, user *entity.User) (int64, error)

	GetUserId(tx *sql.Tx, user *entity.User) (int64, error)

	DeleteUser(tx *sql.Tx, user *entity.User) error

	// Добавляем чат в БД
	AddChat(tx *sql.Tx, chat *entity.Chat) (int64, error)

	// Обновляет данные о чате в БД
	UpdateChat(tx *sql.Tx, chat *entity.Chat) (int64, error)

	// Возваращает текущие параметры для указанной группы
	GetChat(tx *sql.Tx, TgId int64) (entity.Chat, error)

	GetChatId(tx *sql.Tx, chat *entity.Chat) (int64, error)

	DeleteChat(tx *sql.Tx, chat *entity.Chat) error

	AddUserInChat(tx *sql.Tx, userId, chatID int64) (int64, error)

	DeleteUserFromChat(tx *sql.Tx, userTgId, chatTgId int64) error

	AddEvent(tx *sql.Tx, event entity.ChatEvent) (int64, error)

	GetEvent(tx *sql.Tx, event entity.ChatEvent) (int64, error)

	UpdateEvent(tx *sql.Tx, event entity.ChatEvent) (int64, error)

	DeleteEvent(tx *sql.Tx, event entity.ChatEvent) error

	// @eventId - db uniq id
	// @chatId  - db uniq id
	AddEventInChat(tx *sql.Tx, eventId, chatId int64) (int64, error)

	// @eventId - db uniq id
	// @chatId  - db uniq id
	DeleteEventFromChat(tx *sql.Tx, eventId, chatId int64) error

	// @cronEventId - cron event uniq id
	// @chatId  	- tg uniq id
	DeleteEventFromChatByExternalId(tx *sql.Tx, cronEventId, chatId int64) error

	// Получить список групп пользователя
	GetUserGroups(tx *sql.Tx, user *entity.User) ([]entity.Chat, error)

	// Возвращает список всех юзеров состоящих в указанной группе
	GetUsersFromChat(tx *sql.Tx, TgId int64) ([]entity.User, error)

	GetAllChats(tx *sql.Tx) ([]entity.Chat, error)

	IsAdmin(tx *sql.Tx, userTgId int64) (bool, error)
}

type repoImpl struct {
	pg *psql.Postgres
}

func (r *repoImpl) BeginTx() (*sql.Tx, error) {
	return r.pg.BeginTx()
}

func (r *repoImpl) CommitTx(tx *sql.Tx) error {
	return r.pg.CommitTx(tx)
}

func (r *repoImpl) RollbackTx(tx *sql.Tx) error {
	return r.pg.RollbackTx(tx)
}

// Добавляет юзера в БД. Если такой юзер уже есть, обновляет данные в БД
func (r *repoImpl) AddUser(tx *sql.Tx, user *entity.User) (int64, error) {
	if user == nil || tx == nil {
		return -1, errors.New("invalid input parameter")
	}

	// Проверяем что юзера нет в БД
	_, err := r.pg.GetUserByTgID(tx, user.TgID)
	if err != nil && err != sql.ErrNoRows {
		return -1, err
	}

	if err == nil {
		return r.UpdateUser(tx, user)
	}

	// Добавляем юзера
	userId, err := r.pg.AddUser(tx, *user)
	if err != nil {
		return -1, err
	}

	return userId, err
}

// Обновляет данные юзера в БД
func (r *repoImpl) UpdateUser(tx *sql.Tx, user *entity.User) (int64, error) {
	if user == nil || tx == nil {
		return -1, errors.New("invalid input parameter")
	}

	_, err := r.pg.GetUserByTgID(tx, user.TgID)
	if err != nil {
		return -1, err
	}

	id, err := r.pg.UpdateUser(tx, *user)
	if err != nil {
		return -1, err
	}

	return id, err
}

// Возвращает ID юзера из БД
func (r *repoImpl) GetUserId(tx *sql.Tx, user *entity.User) (int64, error) {
	if user == nil || tx == nil {
		return -1, errors.New("invalid input parameter")
	}

	userID, err := r.pg.GetUserIdByTgID(tx, user.TgID)
	if err != nil {
		return -1, err
	}

	return userID, err
}

// Удаляет юзера из бд
func (r *repoImpl) DeleteUser(tx *sql.Tx, user *entity.User) error {
	if user == nil || tx == nil {
		return errors.New("invalid input parameter")
	}

	err := r.pg.DeleteUser(tx, user.TgID)
	if err != nil {
		return err
	}

	return err
}

// Добавляет чат в БД и все его производные таблицы
// Если такой чат уже есть, обновляет данные в таблицах
func (r *repoImpl) AddChat(tx *sql.Tx, chat *entity.Chat) (int64, error) {
	if chat == nil || tx == nil {
		return -1, errors.New("invalid input parameter")
	}

	// Проверяем что такой чат еще не существует
	chatDTO, err := r.pg.GetChat(tx, chat.TgID)
	if err != nil && err != sql.ErrNoRows {
		return -1, err
	}
	if err == sql.ErrNoRows {
		return r.UpdateChat(tx, chat)
	}

	// Добавляем чат в основную таблицу
	chatDTO.Title = chat.Title
	chatDTO.Type = chat.Type
	chatDTO.TgID = chat.TgID

	chatDTO.LastPesChoice = chat.LastPesChoice
	chatDTO.LastCatChoice = chat.LastCatChoice

	// Проверяем что все юзеры из чата есть в БД
	if chat.LastPes != nil {
		chatDTO.LastPesID, err = r.pg.AddUser(tx, *chat.LastPes)
		if err != nil {
			return -1, err
		}
	}
	if chat.LastCat != nil {
		chatDTO.LastCatID, err = r.pg.AddUser(tx, *chat.LastCat)
		if err != nil {
			return -1, err
		}
	}

	chatDTO.OpPerTimeLimiterID, err = r.pg.AddLimiter(tx, psql.LimiterDTO{
		ID:     -1,
		Limit:  float64(chat.OpPerTime.Limit()),
		Burst:  chat.OpPerTime.Burst(),
		Tokens: chat.OpPerTime.Tokens(),
	})
	if err != nil {
		return -1, err
	}

	chatDTO.LuckyCatLimiterID, err = r.pg.AddLimiter(tx, psql.LimiterDTO{
		ID:     -1,
		Limit:  float64(chat.LuckyCatLimiter.Limit()),
		Burst:  chat.LuckyCatLimiter.Burst(),
		Tokens: chat.LuckyCatLimiter.Tokens(),
	})
	if err != nil {
		return -1, err
	}

	chatDTO.LuckyPesLimiterID, err = r.pg.AddLimiter(tx, psql.LimiterDTO{
		ID:     -1,
		Limit:  float64(chat.LuckyPesLimiter.Limit()),
		Burst:  chat.LuckyPesLimiter.Burst(),
		Tokens: chat.LuckyPesLimiter.Tokens(),
	})
	if err != nil {
		return -1, err
	}

	chatId, err := r.pg.AddChat(tx, chatDTO)
	if err != nil {
		return -1, err
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

	return chatId, err
}

func (r *repoImpl) UpdateChat(tx *sql.Tx, chat *entity.Chat) (int64, error) {
	if chat == nil || tx == nil {
		return -1, errors.New("invalid input parameter")
	}

	// Проверяем что такой чат еще не существует
	chatDTO, err := r.pg.GetChat(tx, chat.TgID)
	if err != nil && err != sql.ErrNoRows {
		return -1, err
	}
	if err == sql.ErrNoRows {
		return r.AddChat(tx, chat)
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
		return -1, err
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

	return chatId, err
}

func (r *repoImpl) GetChatByTgId(tx *sql.Tx, chatId int64) (entity.Chat, error) {
	if tx == nil || chatId < 0 {
		return entity.Chat{}, errors.New("invalid input parameter")
	}

	chatDTO, err := r.pg.GetChat(tx, chatId)
	if err != nil {
		return entity.Chat{}, err
	}

	lastCat, err := r.pg.GetUserByTgID(tx, chatDTO.LastCatID)
	if err != nil {
		return entity.Chat{}, err
	}

	lastPes, err := r.pg.GetUserByTgID(tx, chatDTO.LastPesID)
	if err != nil {
		return entity.Chat{}, err
	}

	users, err := r.pg.GetChatMembersByGroupId(tx, chatDTO.Id)
	if err != nil {
		return entity.Chat{}, err
	}

	eventsDTO, err := r.pg.GetChatEvents(tx, chatId)
	if err != nil {
		return entity.Chat{}, err
	}
	var events []entity.ChatEvent
	for _, event := range eventsDTO {
		events = append(events, entity.ChatEvent{
			CronID:     event.CronID,
			TgID:       event.TgID,
			Title:      event.Title,
			Message:    event.Message,
			TimeConfig: event.TimeConfig,
		})
	}

	opLimiterDTO, err := r.pg.GetLimiterByID(tx, chatDTO.OpPerTimeLimiterID)
	if err != nil {
		return entity.Chat{}, err
	}
	opLimiter := rate.Limiter{}
	opLimiter.SetLimit(rate.Limit(opLimiterDTO.Limit))
	opLimiter.SetBurst(opLimiterDTO.Burst)

	luckyCatLimiterDTO, err := r.pg.GetLimiterByID(tx, chatDTO.LuckyCatLimiterID)
	if err != nil {
		return entity.Chat{}, err
	}
	luckyCatLimiter := rate.Limiter{}
	luckyCatLimiter.SetLimit(rate.Limit(luckyCatLimiterDTO.Limit))
	luckyCatLimiter.SetBurst(luckyCatLimiterDTO.Burst)

	luckyPesLimiterDTO, err := r.pg.GetLimiterByID(tx, chatDTO.LuckyPesLimiterID)
	if err != nil {
		return entity.Chat{}, err
	}
	luckyPesLimiter := rate.Limiter{}
	luckyPesLimiter.SetLimit(rate.Limit(luckyPesLimiterDTO.Limit))
	luckyPesLimiter.SetBurst(luckyPesLimiterDTO.Burst)

	return entity.Chat{
		TgID:            chatDTO.TgID,
		Title:           chatDTO.Title,
		Type:            chatDTO.Type,
		LastCat:         lastCat,
		LastPes:         lastPes,
		LastCatChoice:   chatDTO.LastCatChoice,
		LastPesChoice:   chatDTO.LastPesChoice,
		OpPerTime:       &opLimiter,
		LuckyCatLimiter: &luckyCatLimiter,
		LuckyPesLimiter: &luckyPesLimiter,
		Members:         users,
		Events:          events,
	}, err
}

func (r *repoImpl) DeleteChat(tx *sql.Tx, chat *entity.Chat) error {
	if chat == nil || tx == nil {
		return errors.New("invalid input parameter")
	}

	// Удаляем ивенты

	// Удаляем ивенты чата

	// Удаляем лимитеры чата

	// Удаляем мемберов чата

	// Удаляем сам чат
	err = r.pg.DeleteChat(tx, *chat)
	if err != nil {
		r.pg.RollbackTx(tx)
		return err
	}

	return err
}

func (r *repoImpl) AddUserInChat(tx *sql.Tx, userId, chatID int64) (int64, error) {
	if tx == nil || userId < 0 || chatID < 0 {
		return -1, errors.New("invalid input parameter")
	}

}

func (r *repoImpl) DeleteUserFromChat(tx *sql.Tx, userTgId, chatTgId int64) error {
	if tx == nil || userTgId <= 0 || chatTgId <= 0 {
		return errors.New("invalid input parameter")
	}

	//TODO implement me
	panic("implement me")
}

func (r *repoImpl) AddEvent(tx *sql.Tx, event *entity.ChatEvent) (int64, error) {
	if tx == nil || event == nil {
		return -1, errors.New("invalid input parameter")
	}

	//TODO implement me
	panic("implement me")
}

func (r *repoImpl) GetEvent(tx *sql.Tx, event *entity.ChatEvent) (int64, error) {
	if tx == nil || event == nil {
		return -1, errors.New("invalid input parameter")
	}

	//TODO implement me
	panic("implement me")
}

func (r *repoImpl) UpdateEvent(tx *sql.Tx, event *entity.ChatEvent) (int64, error) {
	if tx == nil || event == nil {
		return -1, errors.New("invalid input parameter")
	}

	//TODO implement me
	panic("implement me")
}

func (r *repoImpl) DeleteEvent(tx *sql.Tx, event *entity.ChatEvent) error {
	if tx == nil || event == nil {
		return errors.New("invalid input parameter")
	}

	//TODO implement me
	panic("implement me")
}

func (r *repoImpl) AddEventInChat(tx *sql.Tx, eventId, chatId int64) (int64, error) {
	if tx == nil || eventId < 0 || chatId < 0 {
		return -1, errors.New("invalid input parameter")
	}

	//TODO implement me
	panic("implement me")
}

func (r *repoImpl) DeleteEventFromChat(tx *sql.Tx, eventId, chatId int64) error {
	if tx == nil || eventId < 0 || chatId < 0 {
		return errors.New("invalid input parameter")
	}

	//TODO implement me
	panic("implement me")
}

func (r *repoImpl) GetUserGroups(tx *sql.Tx, user *entity.User) ([]entity.Chat, error) {
	if tx == nil || user == nil {
		return []entity.Chat{}, errors.New("invalid input parameter")
	}

	//TODO implement me
	panic("implement me")
}

func (r *repoImpl) GetAllChats(tx *sql.Tx) ([]entity.Chat, error) {
	if tx == nil {
		return []entity.Chat{}, errors.New("invalid input parameter")
	}

	//TODO implement me
	panic("implement me")
}

func (r *repoImpl) IsAdmin(tx *sql.Tx, userTgId int64) (bool, error) {
	if tx == nil || userTgId <= 0 {
		return false, errors.New("invalid input parameter")
	}

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
