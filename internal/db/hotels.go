package db

import (
	"context"
	"fmt"
	"strings"

	"github.com/stdpmk/hotels/internal/models"
)

func (db *DB) GetHotels(ctx context.Context, f models.HotelsFilter) ([]models.Hotel, int, error) {
	args := []any{}
	where := []string{}
	n := 1

	if f.City != "" {
		where = append(where, fmt.Sprintf("h.city ILIKE $%d", n))
		args = append(args, "%"+f.City+"%")
		n++
	}

	if f.Guests > 0 {
		where = append(where, fmt.Sprintf(
			"EXISTS (SELECT 1 FROM rooms r WHERE r.hotel_id = h.id AND r.max_guests >= $%d)", n,
		))
		args = append(args, f.Guests)
		n++
	}

	if f.CheckIn != nil && f.CheckOut != nil {
		where = append(where, fmt.Sprintf(`EXISTS (
			SELECT 1 FROM rooms r WHERE r.hotel_id = h.id
			AND NOT EXISTS (
				SELECT 1 FROM bookings b WHERE b.room_id = r.id
				AND b.status != $%d
				AND b.check_out > $%d AND b.check_in < $%d
			)
		)`, n, n+1, n+2))
		args = append(args, models.BookingStatusCancelled, *f.CheckIn, *f.CheckOut)
		n += 3
	}

	cond := "1=1"
	if len(where) > 0 {
		cond = strings.Join(where, " AND ")
	}

	var total int
	countQuery := fmt.Sprintf("SELECT COUNT(DISTINCT h.id) FROM hotels h WHERE %s", cond)
	if err := db.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	limit := f.Limit
	if limit <= 0 {
		limit = 20
	}
	page := f.Page
	if page <= 0 {
		page = 1
	}
	offset := (page - 1) * limit

	args = append(args, limit, offset)
	query := fmt.Sprintf(`
		SELECT DISTINCT h.id, h.name, h.city, h.address, h.description, h.rating, h.stars
		FROM hotels h
		WHERE %s
		ORDER BY h.id
		LIMIT $%d OFFSET $%d`, cond, n, n+1)

	rows, err := db.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var result []models.Hotel
	for rows.Next() {
		var hotel models.Hotel
		if err := rows.Scan(&hotel.ID, &hotel.Name, &hotel.City, &hotel.Address, &hotel.Description, &hotel.Rating, &hotel.Stars); err != nil {
			return nil, 0, err
		}
		result = append(result, hotel)
	}
	if rows.Err() != nil {
		return nil, 0, rows.Err()
	}

	return result, total, nil
}

func (db *DB) GetHotelByID(ctx context.Context, id string) (models.Hotel, error) {
	var hotel models.Hotel
	err := db.db.QueryRowContext(ctx,
		`SELECT id, name, city, address, description, rating, stars FROM hotels WHERE id = $1`, id,
	).Scan(&hotel.ID, &hotel.Name, &hotel.City, &hotel.Address, &hotel.Description, &hotel.Rating, &hotel.Stars)
	if err != nil {
		return models.Hotel{}, err
	}
	return hotel, nil
}
