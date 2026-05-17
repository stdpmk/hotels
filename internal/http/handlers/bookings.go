package handlers

import (
	"context"
	"net/http"

	"github.com/stdpmk/hotels/internal/http/middleware"
	"github.com/stdpmk/hotels/internal/http/response"
	"github.com/stdpmk/hotels/internal/models"
)

type bookingsService interface {
	GetMyBookings(ctx context.Context, userID int64) ([]models.Booking, error)
}

type BookingsHandler struct {
	svc bookingsService
}

func NewBookingsHandler(svc bookingsService) *BookingsHandler {
	return &BookingsHandler{svc: svc}
}

func (h *BookingsHandler) GetMyBookings(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(int64)
	if !ok {
		response.WriteError(w, http.StatusUnauthorized, "unauthorized", response.CodeUnauthorized)
		return
	}

	bookings, err := h.svc.GetMyBookings(r.Context(), userID)
	if err != nil {
		response.WriteError(w, http.StatusInternalServerError, "internal server error", response.CodeInternal)
		return
	}

	if bookings == nil {
		bookings = []models.Booking{}
	}

	response.WriteJSON(w, http.StatusOK, bookings)
}
