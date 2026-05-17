-- psql -U aleksandrp -d hotels -f fixtures/users.sql
-- password_hash — bcrypt от 'password123'

INSERT INTO users (email, password_hash, name) VALUES
    ('ivan@example.com', '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy', 'Иван Петров'),
    ('maria@example.com', '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy', 'Мария Иванова'),
    ('alex@example.com', '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy', 'Алексей Сидоров');
