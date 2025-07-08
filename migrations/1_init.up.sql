CREATE TABLE users (
                       id INTEGER PRIMARY KEY UNIQUE ,
                       first_name varchar(80),
                       last_name varchar(80),
                       user_name varchar(80),
                        last_login DATE,
                       photo_url VARCHAR(300),
    is_admin BOOLEAN NOT NULL DEFAULT FALSE
)