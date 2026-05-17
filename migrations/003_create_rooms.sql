CREATE TABLE rooms (
    id               BIGSERIAL PRIMARY KEY,
    hotel_id         BIGINT NOT NULL REFERENCES hotels(id) ON DELETE CASCADE,
    type             TEXT NOT NULL,
    price_per_night  NUMERIC(10,2) NOT NULL,
    max_guests       SMALLINT NOT NULL,
    allow_children   BOOLEAN NOT NULL DEFAULT TRUE,
    allow_pets       BOOLEAN NOT NULL DEFAULT FALSE
);
