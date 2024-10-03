package postgres

import (
	"database/sql"
	"errors"
	"fmt"
	"superserver/entity"
)

/*
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    tg_id BIGINT NOT NULL UNIQUE,
    is_bot BOOLEAN,
    first_name VARCHAR(255),
    last_name VARCHAR(255),
    user_name VARCHAR(255)
);

CREATE TABLE groups (
    id SERIAL PRIMARY KEY,
    tg_id BIGINT NOT NULL UNIQUE,
    title VARCHAR(255),
    type VARCHAR(255),
    last_cat BIGINT,
    last_cat_choise TIMESTAMPTZ,
    last_pes BIGINT,
    last_pes_choise TIMESTAMPTZ,
    last_press_button_lucky_cat TIMESTAMPTZ,
    last_press_button_lucky_pes TIMESTAMPTZ
);

CREATE TABLE members (
    id SERIAL PRIMARY KEY,
    group_id INT REFERENCES groups(id) ON DELETE CASCADE,
    user_id INT REFERENCES users(id) ON DELETE CASCADE,
    UNIQUE(group_id, user_id)
);
*/

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

const AddUserQuery = `
INSERT INTO users (tg_id, is_bot, first_name, last_name, user_name) VALUES ($1, $2, $3, $4, $5) RETURNING id;
`

func (p *Postgres) AddUser(user entity.User, chat entity.Chat) (int64, error) {
	row := p.db.QueryRow(AddUserQuery, user.ID, user.IsBot, user.FirstName, user.LastName, user.UserName)
	var id int64
	err := row.Scan(id)

	return id, err
}

const GetUserByIdQuery = `SELECT * FROM users WHERE tg_id = $1;`

func (p *Postgres) GetUserByTgID(id int64) (entity.User, error) {
	row := p.db.QueryRow(GetUserByIdQuery, id)
	user := entity.User{}

	err := row.Scan(
		&user.ID,
		&user.IsBot,
		&user.FirstName,
		&user.LastName,
		&user.UserName,
	)
	return user, err
}

const UpdateUserQuery = `
UPDATE users
SET first_name = $1, last_name = $2, user_name = $3, is_bot = $4, tg_id = $5
WHERE tg_id = $4 RETURNING id;
`

func (p *Postgres) UpdateUser(user entity.User, chat entity.Chat) (int64, error) {
	row := p.db.QueryRow(UpdateUserQuery, user.FirstName, user.LastName, user.UserName, user.IsBot, user.ID)
	var id int64
	err := row.Scan(id)

	return id, err
}

const DeleteUserQuery = `DELETE FROM users WHERE tg_id = $1;`

func (p *Postgres) DeleteUser(user entity.User) error {
	_, err := p.db.Exec(DeleteUserQuery, user.ID)
	return err
}

const AddChatQuery = `
INSERT INTO groups (tg_id, title, type, last_cat, last_cat_choise, last_pes, last_pes_choise, last_press_button_lucky_cat, last_press_button_lucky_pes)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9) RETURNING id; 
`

func (p *Postgres) AddChat(chat entity.Chat) (int64, error) {
	row := p.db.QueryRow(AddChatQuery, chat.ID, chat.Title, chat.Type, chat.LastCat, chat.LastCatChoice,
		chat.LastPes, chat.LastPesChoice, chat.LastPressButtonLuckyCat, chat.LastPressButtonLuckyPes)
	var id int64
	err := row.Scan(id)
	return id, err
}

const UpdateChatQuery = `
UPDATE groups
SET title = $1, type = $2, last_cat = $3, last_cat_choise = $4, last_pes = $5, last_pes_choise = $6, last_press_button_lucky_cat = $7, last_press_button_lucky_pes = $8
WHERE tg_id = $9 RETURNING id;
`

func (p *Postgres) UpdateChat(chat entity.Chat) (int64, error) {
	row := p.db.QueryRow(UpdateChatQuery, chat.Title, chat.Type, chat.LastCat, chat.LastCatChoice,
		chat.LastPes, chat.LastPesChoice, chat.LastPressButtonLuckyCat, chat.LastPressButtonLuckyPes)

	var id int64
	err := row.Scan(id)

	return id, err
}

const DeleteChatQuery = `DELETE FROM groups WHERE tg_id = $1;`

func (p *Postgres) DeleteChat(chat entity.Chat) error {
	_, err := p.db.Exec(DeleteChatQuery, chat.ID)
	return err
}

const DeleteUserFromChatQuery = `
DELETE FROM members
USING groups, users
WHERE members.group_id = groups.id
AND members.user_id = users.id
AND groups.tg_id = $1
AND users.tg_id = $2;
`

func (p *Postgres) DeleteUserFromChat(user entity.User, chat entity.Chat) error {
	_, err := p.db.Exec(DeleteUserFromChatQuery, user.ID, chat.ID)
	return err
}

const AddUserInChatQuery = `
INSERT INTO members (group_id, user_id)
SELECT g.id, u.id
FROM groups g, users u
WHERE g.tg_id = $1
AND u.tg_id = $2 RETURNING id;
`

func (p *Postgres) AddUserInChat(user entity.User, chat entity.Chat) (int64, error) {
	row := p.db.QueryRow(AddUserInChatQuery, user.ID, chat.ID)

	var id int64
	err := row.Scan(&id)

	return id, err
}

/*
SELECT g.*
FROM groups g
JOIN members m ON g.id = m.group_id
JOIN users u ON m.user_id = u.id
WHERE u.tg_id = $1;
*/
const GetUserChatsQuery = `
SELECT g.*
FROM groups g
JOIN members m ON g.id = m.group_id
JOIN users u ON m.user_id = u.id
WHERE u.tg_id = $1;
`

func (p *Postgres) GetUserChats(user entity.User, chat entity.Chat) ([]entity.Chat, error) {
row:
	-p.db.QueryRow(GetUserChatsQuery)
	return nil, nil
}

/*
SELECT u.*
FROM users u
JOIN members m ON u.id = m.user_id
JOIN groups g ON m.group_id = g.id
WHERE g.tg_id = $1;
*/
func (p *Postgres) GetChatUsers(pq *sql.DB, user entity.User, chat entity.Chat) ([]entity.User, error) {
	return nil, nil
}

/*
SELECT title
FROM groups
WHERE tg_id = $1;
*/
func (p *Postgres) GetChatTitle(pq *sql.DB, chat entity.Chat) (string, error) {
	return "", nil
}
