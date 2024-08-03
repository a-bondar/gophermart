BEGIN TRANSACTION;

CREATE TYPE order_status AS ENUM ('NEW', 'PROCESSING', 'INVALID', 'PROCESSED');

CREATE TABLE IF NOT EXISTS orders
(
    id           INT PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    user_id      INT                      NOT NULL,
    order_number VARCHAR(255) UNIQUE      NOT NULL,
    accrual      NUMERIC(10, 2)           NOT NULL DEFAULT 0 CHECK (accrual >= 0),
    status       order_status             NOT NULL DEFAULT 'NEW',
    uploaded_at  TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users (id)
);

COMMIT;