CREATE TABLE IF NOT EXISTS veterinarians
(
    id           UUID PRIMARY KEY,
    name         TEXT   NOT NULL,
    address      TEXT   NOT NULL,
    customer_ids UUID[] NOT NULL
)
