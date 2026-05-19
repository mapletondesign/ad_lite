package bookings

import "time"

type Booking struct {
	ID           string    `json:"id"`
	SlotID       string    `json:"slot_id"`
	AdvertiserID string    `json:"advertiser_id"`
	CreativeURL  *string   `json:"creative_url,omitempty"`
	StartsOn     string    `json:"starts_on"` // "YYYY-MM-DD"
	EndsOn       string    `json:"ends_on"`
	PriceCents   int       `json:"price_cents"`
	Status       string    `json:"status"` // pending | active | completed | cancelled
	CreatedAt    time.Time `json:"created_at"`
}

type CreateRequest struct {
	SlotID       string  `json:"slot_id"`
	AdvertiserID string  `json:"advertiser_id"`
	CreativeURL  *string `json:"creative_url,omitempty"`
	StartsOn     string  `json:"starts_on"`
	EndsOn       string  `json:"ends_on"`
}

type ListFilter struct {
	SlotID       string
	AdvertiserID string
	Status       string
}
