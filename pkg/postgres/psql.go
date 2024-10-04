package postgres

import (
	"database/sql"
	"errors"
	"fmt"
	_ "github.com/lib/pq"
	"superserver/entity"
	"time"
)

type Postgres struct {
	db *sql.DB
}

func NewPostgres() (*Postgres, error) {
	postgres := &Postgres{}
	connectionParameters := "user=postgres password=postgres dbname=telegram host=localhost sslmode=disable"

	var err error
	postgres.db, err = sql.Open("postgres", connectionParameters)
	if err != nil {
		return nil, err
	}

	err = postgres.db.Ping()
	if err != nil {
		return nil, errors.New(fmt.Sprintf("ошибка подключения к БД: %s", err.Error()))
	}

	return postgres, nil
}

func (p *Postgres) BeginTx() (*sql.Tx, error) {
	return p.db.Begin()
}

func (p *Postgres) CommitTx(tx *sql.Tx) error {
	return tx.Commit()
}

func (p *Postgres) RollbackTx(tx *sql.Tx) error {
	return tx.Rollback()
}

func (p *Postgres) Close() {
	p.db.Close()
}

// User CRUD

const AddUserQuery = `
INSERT INTO users (tg_id, is_bot, first_name, last_name, user_name, language_code, can_join_groups, can_read_all_group_messages, supports_inline_queries)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
RETURNING id;
`

func (p *Postgres) AddUser(tx *sql.Tx, user entity.User) (int64, error) {
	var userID int64
	err := tx.QueryRow(AddUserQuery, user.TgID, user.IsBot, user.FirstName, user.LastName, user.UserName,
		user.LanguageCode, user.CanJoinGroups, user.CanReadAllGroupMessages, user.SupportsInlineQueries).Scan(&userID)
	if err != nil {
		return -1, err
	}
	return userID, nil
}

const GetUserByTgIDQuery = `
SELECT is_bot, first_name, last_name, user_name, language_code, can_join_groups, 
       can_read_all_group_messages, supports_inline_queries
FROM users
WHERE tg_id = $1;
`

func (p *Postgres) GetUserByTgID(tx *sql.Tx, tgID int64) (entity.User, error) {
	var user entity.User
	err := tx.QueryRow(GetUserByTgIDQuery, tgID).Scan(
		&user.IsBot, &user.FirstName, &user.LastName, &user.UserName,
		&user.LanguageCode, &user.CanJoinGroups, &user.CanReadAllGroupMessages, &user.SupportsInlineQueries,
	)

	user.TgID = tgID
	return user, err
}

const UpdateUserQuery = `
UPDATE users
SET is_bot = $1, first_name = $2, last_name = $3, user_name = $4, language_code = $5, can_join_groups = $6, 
    can_read_all_group_messages = $7, supports_inline_queries = $8
WHERE tg_id = $9;
`

func (p *Postgres) UpdateUser(tx *sql.Tx, user entity.User) error {
	_, err := tx.Exec(UpdateUserQuery, user.IsBot, user.FirstName, user.LastName, user.UserName, user.LanguageCode,
		user.CanJoinGroups, user.CanReadAllGroupMessages, user.SupportsInlineQueries, user.TgID)
	return err
}

const DeleteUserQuery = `
DELETE FROM users
WHERE tg_id = $1;
`

func (p *Postgres) DeleteUser(tx *sql.Tx, tgID int64) error {
	_, err := tx.Exec(DeleteUserQuery, tgID)
	return err
}

// Chat CRUD

type ChatDTO struct {
	TgID               int64
	Title              string
	Type               string
	LastCatID          int64
	LastPesID          int64
	LastCatChoice      time.Time
	LastPesChoice      time.Time
	OpPerTimeLimiterID int64
	LuckyCatLimiterID  int64
	LuckyPesLimiterID  int64
	MembersID          []int64
	EventsID           []int64
}

const AddGroupQuery = `
INSERT INTO groups (tg_id, title, type, last_cat, last_cat_choise, last_pes, last_pes_choise, op_per_time_limeter_id, 
                    lucky_cat_limiter_id, lucky_pes_limiter_id)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
RETURNING id;
`

func (p *Postgres) AddChat(tx *sql.Tx, chat ChatDTO) (int64, error) {
	row := tx.QueryRow(AddGroupQuery, chat.TgID, chat.Title, chat.Type, chat.LastCatID, chat.LastCatChoice,
		chat.LastPesID, chat.LastPesChoice, chat.OpPerTimeLimiterID, chat.LuckyCatLimiterID, chat.LuckyPesLimiterID)
	var id int64
	err := row.Scan(id)
	return id, err
}

const GetGroupByTgIDQuery = `
SELECT tg_id, title, type, last_cat, last_cat_choise, last_pes, last_pes_choise, op_per_time_limeter_id, 
       lucky_cat_limiter_id, lucky_pes_limiter_id
FROM groups
WHERE tg_id = $1;
`

func (p *Postgres) GetChat(tx *sql.Tx, tg_id int64) (ChatDTO, error) {
	var chat ChatDTO
	err := tx.QueryRow(GetGroupByTgIDQuery, tg_id).Scan(
		&chat.TgID, &chat.Title, &chat.Type, &chat.LastCatID, &chat.LastCatChoice,
		&chat.LastPesID, &chat.LastPesChoice, &chat.OpPerTimeLimiterID, &chat.LuckyCatLimiterID, &chat.LuckyPesLimiterID,
	)

	return chat, err
}

const UpdateGroupQuery = `
UPDATE groups
SET title = $1, type = $2, last_cat = $3, last_cat_choise = $4, last_pes = $5, last_pes_choise = $6, 
    op_per_time_limeter_id = $7, lucky_cat_limiter_id = $8, lucky_pes_limiter_id = $9
WHERE tg_id = $10;
`

func (p *Postgres) UpdateChat(tx *sql.Tx, chat ChatDTO) error {
	_, err := tx.Exec(UpdateGroupQuery, chat.Title, chat.Type, chat.LastCatID, chat.LastCatChoice, chat.LastPesID,
		chat.LastPesChoice, chat.OpPerTimeLimiterID, chat.LuckyCatLimiterID, chat.LuckyPesLimiterID, chat.TgID)

	return err
}

const DeleteChatQuery = `DELETE FROM groups WHERE tg_id = $1;`

func (p *Postgres) DeleteChat(tx *sql.Tx, chat entity.Chat) error {
	_, err := tx.Exec(DeleteChatQuery, chat.TgID)
	return err
}

// Limiter CRUD

type LimiterDTO struct {
	ID        int64
	Limit     float64
	Burst     int
	Tokens    float64
	Last      time.Time
	LastEvent time.Time
}

const AddLimiterQuery = `
INSERT INTO limiters (burst, limit, tokens, last, last_event)
VALUES ($1, $2, $3, $4, $5)
RETURNING id;
`

func (p *Postgres) AddLimiter(tx *sql.Tx, limiter LimiterDTO) (int64, error) {
	var limiterID int64
	err := tx.QueryRow(AddLimiterQuery, limiter.Burst, limiter.Limit, limiter.Tokens, limiter.Last, limiter.LastEvent).Scan(&limiterID)

	return limiterID, err
}

const GetLimiterByIDQuery = `
SELECT burst, limit, tokens, last, last_event
FROM limiters
WHERE id = $1;
`

func (p *Postgres) GetLimiterByID(tx *sql.Tx, id int64) (LimiterDTO, error) {
	var limiter LimiterDTO
	err := tx.QueryRow(GetLimiterByIDQuery, id).Scan(
		&limiter.Burst, &limiter.Limit, &limiter.Tokens, &limiter.Last, &limiter.LastEvent,
	)

	return limiter, err
}

const UpdateLimiterQuery = `
UPDATE limiters
SET burst = $1, limit = $2, tokens = $3, last = $4, last_event = $5
WHERE id = $6;
`

func (p *Postgres) UpdateLimiter(tx *sql.Tx, limiter LimiterDTO) error {
	_, err := tx.Exec(UpdateLimiterQuery, limiter.Burst, limiter.Limit, limiter.Tokens, limiter.Last, limiter.LastEvent, limiter.ID)

	return err
}

const DeleteLimiterQuery = `
DELETE FROM limiters
WHERE id = $1;
`

func (p *Postgres) DeleteLimiter(tx *sql.Tx, id int64) error {
	_, err := tx.Exec(DeleteLimiterQuery, id)

	return err
}

// Event CRUD

type EventDTO struct {
	ID         int64
	CronID     int64
	TgID       int64
	Title      string
	Message    string
	TimeConfig string
}

const AddEventQuery = `
INSERT INTO events (cron_id, tg_id, title, message, time_config)
VALUES ($1, $2, $3, $4, $5)
RETURNING id;
`

func (p *Postgres) AddEvent(tx *sql.Tx, event EventDTO) (int64, error) {
	var eventID int64
	err := tx.QueryRow(AddEventQuery, event.CronID, event.TgID, event.Title, event.Message, event.TimeConfig).Scan(&eventID)

	return eventID, err
}

const GetEventByIDQuery = `
SELECT cron_id, tg_id, title, message, time_config
FROM events
WHERE id = $1;
`

func (p *Postgres) GetEventByID(tx *sql.Tx, id int64) (EventDTO, error) {
	var event EventDTO
	err := tx.QueryRow(GetEventByIDQuery, id).Scan(
		&event.CronID, &event.TgID, &event.Title, &event.Message, &event.TimeConfig,
	)

	return event, err
}

const UpdateEventQuery = `
UPDATE events
SET cron_id = $1, tg_id = $2, title = $3, message = $4, time_config = $5
WHERE id = $6;
`

func (p *Postgres) UpdateEvent(tx *sql.Tx, event EventDTO) error {
	_, err := tx.Exec(UpdateEventQuery, event.CronID, event.TgID, event.Title, event.Message, event.TimeConfig, event.ID)

	return err
}

const DeleteEventQuery = `
DELETE FROM events
WHERE id = $1;
`

func (p *Postgres) DeleteEvent(tx *sql.Tx, id int64) error {
	_, err := tx.Exec(DeleteEventQuery, id)

	return err
}
