CREATE TABLE users (
                       id varchar(255) PRIMARY KEY UNIQUE ,
    hash varchar(255),
                       first_name varchar(80),
                       last_name varchar(80),
                       user_name varchar(80),
    user_name_locale varchar(80),
                        last_login DATE,
                       photo_url VARCHAR(300),
    is_admin BOOLEAN NOT NULL DEFAULT FALSE
);
CREATE TABLE IF NOT EXISTS apps
(
    id     INTEGER PRIMARY KEY,
    name   TEXT NOT NULL UNIQUE,
    secret TEXT NOT NULL UNIQUE
);
