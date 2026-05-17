package models

import "time"

type Booking struct {
	ID         int64     `json:"id"`
	UserID     int64     `json:"user_id"`
	RoomID     int64     `json:"room_id"`
	CheckIn    time.Time `json:"check_in"`
	CheckOut   time.Time `json:"check_out"`
	TotalPrice float64   `json:"total_price"`
	Status     string    `json:"status"`
	CreatedAt  time.Time `json:"created_at"`
}

type HotelInfo struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
	City string `json:"city"`
}

type RoomInfo struct {
	ID     int64     `json:"id"`
	Number string    `json:"number"`
	Type   string    `json:"type"`
	Hotel  HotelInfo `json:"hotel"`
}

type BookingDetail struct {
	ID         int64     `json:"id"`
	CheckIn    time.Time `json:"check_in"`
	CheckOut   time.Time `json:"check_out"`
	TotalPrice float64   `json:"total_price"`
	Status     string    `json:"status"`
	CreatedAt  time.Time `json:"created_at"`
	Room       RoomInfo  `json:"room"`
}
