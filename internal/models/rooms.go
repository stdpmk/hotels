package models

type Room struct {
	ID            int64      `json:"id"`
	HotelID       int64      `json:"hotel_id"`
	Type          string     `json:"type"`
	PricePerNight float64    `json:"price_per_night"`
	MaxGuests     int        `json:"max_guests"`
	AllowChildren bool       `json:"allow_children"`
	AllowPets     bool       `json:"allow_pets"`
}
