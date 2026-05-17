-- psql -U aleksandrp -d hotels -f fixtures/rooms.sql

-- Марриотт Москва (hotel_id=1)
INSERT INTO rooms (hotel_id, type, price_per_night, max_guests, allow_children, allow_pets) VALUES
    (1, 'Стандарт', 8500.00, 2, true, false),
    (1, 'Делюкс', 12000.00, 2, true, false),
    (1, 'Люкс', 25000.00, 3, true, false),
    (1, 'Президентский люкс', 60000.00, 4, true, false);

-- Хилтон Санкт-Петербург (hotel_id=2)
INSERT INTO rooms (hotel_id, type, price_per_night, max_guests, allow_children, allow_pets) VALUES
    (2, 'Стандарт', 7000.00, 2, true, false),
    (2, 'Делюкс', 10500.00, 2, true, true),
    (2, 'Семейный', 14000.00, 4, true, false);

-- Новотель Казань (hotel_id=3)
INSERT INTO rooms (hotel_id, type, price_per_night, max_guests, allow_children, allow_pets) VALUES
    (3, 'Стандарт', 4500.00, 2, true, false),
    (3, 'Делюкс', 6500.00, 2, true, true),
    (3, 'Семейный', 9000.00, 4, true, false);

-- Ибис Новосибирск (hotel_id=4)
INSERT INTO rooms (hotel_id, type, price_per_night, max_guests, allow_children, allow_pets) VALUES
    (4, 'Стандарт', 2500.00, 2, true, false),
    (4, 'Двухместный', 3200.00, 2, false, false);

-- Courtyard Сочи (hotel_id=5)
INSERT INTO rooms (hotel_id, type, price_per_night, max_guests, allow_children, allow_pets) VALUES
    (5, 'Стандарт с видом на море', 6000.00, 2, true, false),
    (5, 'Делюкс с балконом', 9500.00, 2, true, false),
    (5, 'Семейный с видом на море', 13000.00, 4, true, true);
