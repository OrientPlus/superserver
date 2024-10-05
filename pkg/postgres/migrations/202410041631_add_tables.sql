-- +goose Up
CREATE TABLE users (
                       id 							SERIAL PRIMARY KEY,
                       tg_id 						BIGINT NOT NULL UNIQUE,
                       is_bot 						BOOLEAN NOT NULL,
                       first_name 					VARCHAR(255),
                       last_name 					VARCHAR(255),
                       user_name 					VARCHAR(255) NOT NULL,
                       language_code 				VARCHAR(255),
                       can_join_groups 			    BOOLEAN,
                       can_read_all_group_messages  BOOLEAN,
                       supports_inline_queries  	BOOLEAN
);

CREATE TABLE groups (
                        id 						SERIAL PRIMARY KEY,
                        tg_id 					BIGINT NOT NULL UNIQUE,
                        title 					VARCHAR(255) NOT NULL,
                        type 					VARCHAR(255) NOT NULL,
                        last_cat 				BIGINT,
                        last_cat_choise 		TIMESTAMPTZ,
                        last_pes 				BIGINT,
                        last_pes_choise 		TIMESTAMPTZ,
                        op_per_time_limeter_id 	INT REFERENCES limiters(id) ON DELETE CASCADE,
                        lucky_cat_limiter_id 	INT REFERENCES limiters(id) ON DELETE CASCADE,
                        lucky_pes_limiter_id 	INT REFERENCES limiters(id) ON DELETE CASCADE
);

CREATE TABLE limiters (
                          id 			SERIAL PRIMARY KEY,
                          burst 		INT NOT NULL,
                          lim    		FLOAT8 NOT NULL,
                          tokens 		FLOAT8 NOT NULL
);

CREATE TABLE members (
                         id 		SERIAL PRIMARY KEY,
                         group_id 	INT REFERENCES groups(id) ON DELETE CASCADE,
                         user_id 	INT REFERENCES users(id) ON DELETE CASCADE,
                         UNIQUE(group_id, user_id)
);

CREATE TABLE chat_events (
                            id          SERIAL PRIMARY KEY,
                            chat_id     INT REFERENCES groups(id),
                            event_id    INT REFERENCES events(id)
);

CREATE TABLE events (
                        id          SERIAL PRIMARY KEY,
                        cron_id     BIGINT NOT NULL,
                        tg_id       BIGINT NOT NULL,
                        title       VARCHAR(255) NOT NULL,
                        message     VARCHAR(255) NOT NULL,
                        time_config VARCHAR(255) NOT NULL
);

CREATE TABLE admins (
                        id      SERIAL PRIMARY KEY,
                        tg_id   BIGINT NOT NULL UNIQUE
);

-- +goose Down
DROP TABLE admins;
DROP TABLE events;
DROP TABLE chat_events;
DROP TABLE members;
DROP TABLE limiters;
DROP TABLE groups;
DROP TABLE users;