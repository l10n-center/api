CREATE TABLE IF NOT EXISTS "user" (
    email       VARCHAR(255) PRIMARY KEY,
    passhash    CHAR(60),
    role        INT,
    reset_token CHAR(60),
    created_at  TIME DEFAULT NOW(),
    updated_at  TIME DEFAULT NOW(),
    deleted_at  TIME
);
