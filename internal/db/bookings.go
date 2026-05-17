package db

import (
	"context"
	"time"

	"github.com/stdpmk/hotels/internal/models"
)

func (db *DB) GetBookingsByUserID(ctx context.Context, userID int64) ([]models.Booking, error) {
	rows, err := db.DB.QueryContext(ctx,
		`SELECT id, user_id, room_id, check_in, check_out, total_price, status, created_at
		 FROM bookings WHERE user_id = $1 ORDER BY created_at DESC`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []models.Booking
	for rows.Next() {
		var b models.Booking
		if err := rows.Scan(&b.ID, &b.UserID, &b.RoomID, &b.CheckIn, &b.CheckOut, &b.TotalPrice, &b.Status, &b.CreatedAt); err != nil {
			return nil, err
		}
		result = append(result, b)
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}

	return result, nil
}

func (db *DB) GetBookingDetailsByUserID(ctx context.Context, userID int64) ([]models.BookingDetail, error) {
	rows, err := db.DB.QueryContext(ctx,
		`SELECT
			b.id, b.check_in, b.check_out, b.total_price, b.status, b.created_at,
			r.id, r.number, r.type,
			h.id, h.name, h.city
		 FROM bookings b
		 JOIN rooms r ON r.id = b.room_id
		 JOIN hotels h ON h.id = r.hotel_id
		 WHERE b.user_id = $1
		 ORDER BY b.created_at DESC`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []models.BookingDetail
	for rows.Next() {
		var b models.BookingDetail
		if err := rows.Scan(
			&b.ID, &b.CheckIn, &b.CheckOut, &b.TotalPrice, &b.Status, &b.CreatedAt,
			&b.Room.ID, &b.Room.Number, &b.Room.Type,
			&b.Room.Hotel.ID, &b.Room.Hotel.Name, &b.Room.Hotel.City,
		); err != nil {
			return nil, err
		}
		result = append(result, b)
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}

	return result, nil
}

func (db *DB) IsRoomAvailable(ctx context.Context, roomID int64, checkIn, checkOut time.Time) (bool, error) {
	var count int
	err := db.DB.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM bookings
		 WHERE room_id = $1 AND status != 'cancelled'
		 AND check_out > $2 AND check_in < $3`,
		roomID, checkIn, checkOut,
	).Scan(&count)
	if err != nil {
		return false, err
	}
	return count == 0, nil
}

func (db *DB) CreateBooking(ctx context.Context, userID, roomID int64, checkIn, checkOut time.Time, totalPrice float64) (models.Booking, error) {
	var b models.Booking
	err := db.DB.QueryRowContext(ctx,
		`INSERT INTO bookings (user_id, room_id, check_in, check_out, total_price, status)
		 VALUES ($1, $2, $3, $4, $5, 'confirmed')
		 RETURNING id, user_id, room_id, check_in, check_out, total_price, status, created_at`,
		userID, roomID, checkIn, checkOut, totalPrice,
	).Scan(&b.ID, &b.UserID, &b.RoomID, &b.CheckIn, &b.CheckOut, &b.TotalPrice, &b.Status, &b.CreatedAt)
	if err != nil {
		return models.Booking{}, err
	}
	return b, nil
}
