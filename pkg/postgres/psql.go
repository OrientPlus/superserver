package postgres

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


Создание юзера
INSERT INTO users (tg_id, is_bot, first_name, last_name, user_name)
VALUES (123456789, false, 'John', 'Doe', 'johndoe');


Чтение юзера
SELECT * FROM users WHERE id = 1;


обновление юзера
UPDATE users
SET first_name = 'Jane', last_name = 'Doe', user_name = 'janedoe'
WHERE id = 1;

удаление юзера
DELETE FROM users WHERE id = 1;

создание группы
INSERT INTO groups (tg_id, title, type)
VALUES (987654321, 'My Group', 'public');



чтение группы
SELECT * FROM groups WHERE id = 1;

обновление группы
UPDATE groups
SET title = 'New Group Title', type = 'private'
WHERE id = 1;

удаление групппы
DELETE FROM groups WHERE id = 1;

удаление юзера из группы
DELETE FROM members
USING groups, users
WHERE members.group_id = groups.id
AND members.user_id = users.id
AND groups.tg_id = 987654321
AND users.tg_id = 123456789;

добавление юзера в группу
INSERT INTO members (group_id, user_id)
SELECT g.id, u.id
FROM groups g, users u
WHERE g.tg_id = 987654321
AND u.tg_id = 123456789;


получение списка групп юзера
SELECT g.*
FROM groups g
JOIN members m ON g.id = m.group_id
JOIN users u ON m.user_id = u.id
WHERE u.tg_id = 123456789;

получение списка юзеров в группе
SELECT u.*
FROM users u
JOIN members m ON u.id = m.user_id
JOIN groups g ON m.group_id = g.id
WHERE g.tg_id = 987654321;

получение имени группы
SELECT title
FROM groups
WHERE tg_id = 987654321;








*/
