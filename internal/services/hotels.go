package services

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/stdpmk/hotels/internal/db"
	"github.com/stdpmk/hotels/internal/models"
)

const hotelsKey = "hotels:all"

var ErrHotelNotFound = errors.New("hotel not found")

type HotelsService struct {
	db    *db.DB
	redis *redis.Client
	ttl   time.Duration
}

func NewHotelsService(db *db.DB, redis *redis.Client, ttl time.Duration) *HotelsService {
	return &HotelsService{db: db, redis: redis, ttl: ttl}
}

func (h *HotelsService) GetHotels(ctx context.Context) ([]models.Hotel, error) {
	data, err := h.redis.Get(ctx, hotelsKey).Bytes()
	if err == nil {
		var hotels []models.Hotel
		if err := json.Unmarshal(data, &hotels); err == nil {
			return hotels, nil
		}
	}

	hotels, err := h.db.GetHotels(ctx)
	if err != nil {
		return nil, err
	}

	if data, err := json.Marshal(hotels); err == nil {
		if err := h.redis.Set(ctx, hotelsKey, data, h.ttl).Err(); err != nil {
			log.Printf("cache set error: %v", err)
		}
	}

	return hotels, nil
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

func (h *HotelsService) FindHotelsByFilter(ctx context.Context, filter map[string]any) ([]models.Hotel, error) {
	// TODO: implement me
	return nil, errors.New("Implement me")
}
