CREATE TABLE hotels (
    id          BIGSERIAL PRIMARY KEY,
    name        TEXT NOT NULL,
    city        TEXT NOT NULL,
    address     TEXT,
    description TEXT,
    rating      NUMERIC(2,1),
    stars       SMALLINT CHECK (stars BETWEEN 1 AND 5)
);
