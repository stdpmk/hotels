package services

import (
	"context"
	"database/sql"
	"errors"

	"github.com/stdpmk/hotels/internal/db"
	"github.com/stdpmk/hotels/internal/models"
)

var ErrHotelNotFound = errors.New("hotel not found")

type HotelsService struct {
	db *db.DB
}

func NewHotelsService(db *db.DB) *HotelsService {
	return &HotelsService{db: db}
}

func (h *HotelsService) GetHotels(ctx context.Context, f models.HotelsFilter) (models.HotelsPage, error) {
	hotels, total, err := h.db.GetHotels(ctx, f)
	if err != nil {
		return models.HotelsPage{}, err
	}

	if hotels == nil {
		hotels = []models.Hotel{}
	}

	limit := f.Limit
	if limit <= 0 {
		limit = 20
	}
	page := f.Page
	if page <= 0 {
		page = 1
	}

	return models.HotelsPage{
		Hotels: hotels,
		Total:  total,
		Page:   page,
		Limit:  limit,
	}, nil
}

func (h *HotelsService) GetHotelByID(ctx context.Context, id string) (models.Hotel, error) {
	hotel, err := h.db.GetHotelByID(ctx, id)
	if errors.Is(err, sql.ErrNoRows) {
		return models.Hotel{}, ErrHotelNotFound
	}
	if err != nil {
		return models.Hotel{}, err
	}
	return hotel, nil
}
