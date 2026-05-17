-- psql -U aleksandrp -d hotels -f fixtures/bookings.sql

INSERT INTO bookings (user_id, room_id, check_in, check_out, total_price, status) VALUES
    (1, 1, '2026-06-01', '2026-06-05', 34000.00, 'confirmed'),
    (2, 6, '2026-06-10', '2026-06-13', 21000.00, 'confirmed'),
    (3, 11, '2026-07-01', '2026-07-03', 9000.00, 'pending'),
    (1, 14, '2026-08-15', '2026-08-20', 30000.00, 'pending'),
    (2, 3, '2026-06-20', '2026-06-22', 24000.00, 'cancelled');
