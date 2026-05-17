package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/stdpmk/hotels/internal/http/middleware"
	"github.com/stdpmk/hotels/internal/http/response"
	"github.com/stdpmk/hotels/internal/models"
	"github.com/stdpmk/hotels/internal/services"
)

type BookingsHandler struct {
	svc *services.BookingsService
}

func NewBookingsHandler(svc *services.BookingsService) *BookingsHandler {
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
		bookings = []models.BookingDetail{}
	}

	response.WriteJSON(w, http.StatusOK, bookings)
}

type createBookingRequest struct {
	RoomID   int64  `json:"room_id"`
	CheckIn  string `json:"check_in"`
	CheckOut string `json:"check_out"`
}

func (h *BookingsHandler) CreateBooking(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(int64)
	if !ok {
		response.WriteError(w, http.StatusUnauthorized, "unauthorized", response.CodeUnauthorized)
		return
	}

	var req createBookingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid request body", response.CodeBadRequest)
		return
	}

	checkIn, err := time.Parse("2006-01-02", req.CheckIn)
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid check_in date, use YYYY-MM-DD", response.CodeBadRequest)
		return
	}
	checkOut, err := time.Parse("2006-01-02", req.CheckOut)
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid check_out date, use YYYY-MM-DD", response.CodeBadRequest)
		return
	}

	booking, err := h.svc.CreateBooking(r.Context(), userID, req.RoomID, checkIn, checkOut)
	if errors.Is(err, services.ErrRoomNotFound) {
		response.WriteError(w, http.StatusNotFound, "room not found", response.CodeNotFound)
		return
	}
	if errors.Is(err, services.ErrInvalidDates) {
		response.WriteError(w, http.StatusBadRequest, "check_out must be after check_in", response.CodeBadRequest)
		return
	}
	if errors.Is(err, services.ErrRoomNotAvailable) {
		response.WriteError(w, http.StatusConflict, "room not available for selected dates", response.CodeConflict)
		return
	}
	if err != nil {
		response.WriteError(w, http.StatusInternalServerError, "internal server error", response.CodeInternal)
		return
	}

	response.WriteJSON(w, http.StatusCreated, booking)
}
