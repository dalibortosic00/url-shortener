package models

import "time"

type LinkRecord struct {
	Code      string    `json:"code"`
	URL       string    `json:"url"`
	OwnerID   *string   `json:"owner_id"` // optional for guests, required for registered users
	CreatedAt time.Time `json:"created_at"`
}
