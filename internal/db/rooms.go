package db

import (
	"context"
	"database/sql"

	"github.com/stdpmk/hotels/internal/models"
)

func (db *DB) GetRoomsByHotelID(ctx context.Context, hotelID int64) ([]models.Room, error) {
	rows, err := db.DB.QueryContext(ctx,
		`SELECT id, hotel_id, number, type, price_per_night, max_guests, allow_children, allow_pets
		 FROM rooms WHERE hotel_id = $1 ORDER BY price_per_night`,
		hotelID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []models.Room
	for rows.Next() {
		var r models.Room
		if err := rows.Scan(&r.ID, &r.HotelID, &r.Number, &r.Type, &r.PricePerNight, &r.MaxGuests, &r.AllowChildren, &r.AllowPets); err != nil {
			return nil, err
		}
		result = append(result, r)
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}

	return result, nil
}

func (db *DB) GetRoomByID(ctx context.Context, id int64) (models.Room, error) {
	var r models.Room
	err := db.DB.QueryRowContext(ctx,
		`SELECT id, hotel_id, number, type, price_per_night, max_guests, allow_children, allow_pets
		 FROM rooms WHERE id = $1`,
		id,
	).Scan(&r.ID, &r.HotelID, &r.Number, &r.Type, &r.PricePerNight, &r.MaxGuests, &r.AllowChildren, &r.AllowPets)
	if err == sql.ErrNoRows {
		return models.Room{}, sql.ErrNoRows
	}
	if err != nil {
		return models.Room{}, err
	}
	return r, nil
}
