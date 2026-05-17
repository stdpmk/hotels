package models

import "gopkg.in/guregu/null.v4"

type Hotel struct {
	ID          int64       `json:"id"`
	Name        string      `json:"name"`
	City        string      `json:"city"`
	Address     null.String `json:"address"`
	Description null.String `json:"description"`
	Rating      null.Float  `json:"rating"`
	Stars       null.Int    `json:"stars"`
}
