package services

import (
	"context"

	"github.com/stdpmk/hotels/internal/db"
	"github.com/stdpmk/hotels/internal/models"
)

type BookingsService struct {
	db *db.DB
}

func NewBookingsService(db *db.DB) *BookingsService {
	return &BookingsService{db: db}
}

func (s *BookingsService) GetMyBookings(ctx context.Context, userID int64) ([]models.Booking, error) {
	return s.db.GetBookingsByUserID(ctx, userID)
}
