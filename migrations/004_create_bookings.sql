CREATE TABLE bookings (
    id          BIGSERIAL PRIMARY KEY,
    user_id     BIGINT NOT NULL REFERENCES users(id),
    room_id     BIGINT NOT NULL REFERENCES rooms(id),
    check_in    DATE NOT NULL,
    check_out   DATE NOT NULL,
    total_price NUMERIC(10,2) NOT NULL,
    status      TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'confirmed', 'cancelled')),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT check_dates CHECK (check_out > check_in)
);
