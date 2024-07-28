CREATE TABLE IF NOT EXISTS users
(
    id              INT PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    login           VARCHAR(255)                NOT NULL UNIQUE,
    hashed_password VARCHAR(60)                 NOT NULL,
    created_at      timestamp(0) with time zone NOT NULL DEFAULT NOW()
);
