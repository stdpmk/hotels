-- psql -U aleksandrp -d hotels -f fixtures/001_rooms_sochi.sql

-- Rixos Krasnaya Polyana (hotel_id=29)
INSERT INTO rooms (hotel_id, number, type, price_per_night, max_guests, allow_children, allow_pets) VALUES
    (29, '101', 'Стандарт с видом на горы', 15000.00, 2, true, false),
    (29, '102', 'Стандарт с видом на склон', 17000.00, 2, true, false),
    (29, '201', 'Делюкс с балконом', 24000.00, 2, true, false),
    (29, '202', 'Делюкс с видом на Розу Пик', 28000.00, 2, true, false),
    (29, '301', 'Семейный', 32000.00, 4, true, false),
    (29, '302', 'Семейный с видом на горы', 38000.00, 4, true, false),
    (29, '401', 'Джуниор сюит', 45000.00, 2, true, false),
    (29, '402', 'Джуниор сюит с джакузи', 52000.00, 3, true, false),
    (29, '501', 'Сюит с двумя спальнями', 70000.00, 4, true, false),
    (29, 'P1',  'Президентский сюит', 120000.00, 4, true, false);
