package handlers

import (
	"errors"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/stdpmk/hotels/internal/http/response"
	"github.com/stdpmk/hotels/internal/services"
)

type HotelsHandler struct {
	service *services.HotelsService
}

func NewHotelsHandler(service *services.HotelsService) *HotelsHandler {
	return &HotelsHandler{service: service}
}

func (h *HotelsHandler) GetHotelsHandler(w http.ResponseWriter, req *http.Request) {
	hotels, err := h.service.GetHotels(req.Context())
	if err != nil {
		response.WriteError(w, http.StatusInternalServerError, "internal server error", response.CodeInternal)
		return
	}

	response.WriteJSON(w, http.StatusOK, hotels)
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
