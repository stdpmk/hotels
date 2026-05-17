package models

import "time"

type HotelsFilter struct {
	City     string
	CheckIn  *time.Time
	CheckOut *time.Time
	Guests   int
	Page     int
	Limit    int
}

type HotelsPage struct {
	Hotels []Hotel `json:"hotels"`
	Total  int     `json:"total"`
	Page   int     `json:"page"`
	Limit  int     `json:"limit"`
}
