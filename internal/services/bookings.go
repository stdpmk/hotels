package services

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/stdpmk/hotels/internal/db"
	"github.com/stdpmk/hotels/internal/models"
)

var (
	ErrRoomNotFound     = errors.New("room not found")
	ErrRoomNotAvailable = errors.New("room not available for selected dates")
	ErrInvalidDates     = errors.New("check_out must be after check_in")
)

type BookingsService struct {
	db *db.DB
}

func NewBookingsService(db *db.DB) *BookingsService {
	return &BookingsService{db: db}
}

func (s *BookingsService) GetMyBookings(ctx context.Context, userID int64) ([]models.BookingDetail, error) {
	return s.db.GetBookingDetailsByUserID(ctx, userID)
}

func (s *BookingsService) CreateBooking(ctx context.Context, userID, roomID int64, checkIn, checkOut time.Time) (models.Booking, error) {
	room, err := s.db.GetRoomByID(ctx, roomID)
	if errors.Is(err, sql.ErrNoRows) {
		return models.Booking{}, ErrRoomNotFound
	}
	if err != nil {
		return models.Booking{}, err
	}

	if !checkOut.After(checkIn) {
		return models.Booking{}, ErrInvalidDates
	}

	available, err := s.db.IsRoomAvailable(ctx, roomID, checkIn, checkOut)
	if err != nil {
		return models.Booking{}, err
	}
	if !available {
		return models.Booking{}, ErrRoomNotAvailable
	}

	nights := int(checkOut.Sub(checkIn).Hours() / 24)
	totalPrice := room.PricePerNight * float64(nights)

	return s.db.CreateBooking(ctx, userID, roomID, checkIn, checkOut, totalPrice)
}
