CREATE TABLE IF NOT EXISTS "user" (
    id          SERIAL,
    email       VARCHAR(255),
    passhash    CHAR(60),
    role        INT,
    reset_token CHAR(60),
    created_at  TIME DEFAULT NOW(),
    updated_at  TIME DEFAULT NOW(),
    deleted_at  TIME,

    PRIMARY KEY (id),
    UNIQUE (email)
);
