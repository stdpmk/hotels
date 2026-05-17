package handlers

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/stdpmk/hotels/internal/http/response"
	"github.com/stdpmk/hotels/internal/models"
	"github.com/stdpmk/hotels/internal/services"
)

type HotelsHandler struct {
	service *services.HotelsService
}

func NewHotelsHandler(service *services.HotelsService) *HotelsHandler {
	return &HotelsHandler{service: service}
}

func (h *HotelsHandler) GetHotelsHandler(w http.ResponseWriter, req *http.Request) {
	q := req.URL.Query()

	filter := models.HotelsFilter{
		City:  q.Get("city"),
		Page:  parseInt(q.Get("page"), 1),
		Limit: parseInt(q.Get("limit"), 20),
	}

	if guests := parseInt(q.Get("guests"), 0); guests > 0 {
		filter.Guests = guests
	}

	if ci := q.Get("check_in"); ci != "" {
		if t, err := time.Parse("2006-01-02", ci); err == nil {
			filter.CheckIn = &t
		}
	}
	if co := q.Get("check_out"); co != "" {
		if t, err := time.Parse("2006-01-02", co); err == nil {
			filter.CheckOut = &t
		}
	}

	page, err := h.service.GetHotels(req.Context(), filter)
	if err != nil {
		response.WriteError(w, http.StatusInternalServerError, "internal server error", response.CodeInternal)
		return
	}

	response.WriteJSON(w, http.StatusOK, page)
}

func (h *HotelsHandler) GetHotelByIDHandler(w http.ResponseWriter, req *http.Request) {
	id := mux.Vars(req)["id"]

	hotel, err := h.service.GetHotelByID(req.Context(), id)
	if errors.Is(err, services.ErrHotelNotFound) {
		response.WriteError(w, http.StatusNotFound, "hotel not found", response.CodeNotFound)
		return
	}
	if err != nil {
		response.WriteError(w, http.StatusInternalServerError, "internal server error", response.CodeInternal)
		return
	}

	response.WriteJSON(w, http.StatusOK, hotel)
}

func parseInt(s string, fallback int) int {
	if v, err := strconv.Atoi(s); err == nil && v > 0 {
		return v
	}
	return fallback
}
