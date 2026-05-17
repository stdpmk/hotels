package db

import (
	"context"

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
