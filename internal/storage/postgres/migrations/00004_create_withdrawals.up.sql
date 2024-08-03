CREATE TABLE IF NOT EXISTS withdrawals
(
    id           INT PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    user_id      INT                      NOT NULL,
    sum          NUMERIC(10, 2)           NOT NULL DEFAULT 0 CHECK (sum > 0),
    order_number VARCHAR(255)             NOT NULL,
    processed_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users (id),
    FOREIGN KEY (order_number) REFERENCES orders (order_number)
);