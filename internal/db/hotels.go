package db

import (
	"context"

	"github.com/stdpmk/hotels/internal/models"
)

func (db *DB) GetHotels(ctx context.Context) ([]models.Hotel, error) {
	rows, err := db.DB.QueryContext(ctx, `SELECT id, name, city, address, description, rating, stars FROM hotels`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []models.Hotel
	for rows.Next() {
		var hotel models.Hotel
		if err := rows.Scan(&hotel.ID, &hotel.Name, &hotel.City, &hotel.Address, &hotel.Description, &hotel.Rating, &hotel.Stars); err != nil {
			return nil, err
		}
		result = append(result, hotel)
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}

	return result, nil
}

func (db *DB) GetHotelByID(ctx context.Context, id string) (models.Hotel, error) {
	var hotel models.Hotel
	err := db.DB.QueryRowContext(ctx,
		`SELECT id, name, city, address, description, rating, stars FROM hotels WHERE id = $1`, id,
	).Scan(&hotel.ID, &hotel.Name, &hotel.City, &hotel.Address, &hotel.Description, &hotel.Rating, &hotel.Stars)
	if err != nil {
		return models.Hotel{}, err
	}
	return hotel, nil
}
