package slots

import "time"

type Slot struct {
	ID          string    `json:"id"`
	DeviceID    string    `json:"device_id"`
	Label       *string   `json:"label,omitempty"`
	DaysOfWeek  []int     `json:"days_of_week"` // 0=Sun … 6=Sat
	StartTime   string    `json:"start_time"`   // "HH:MM"
	EndTime     string    `json:"end_time"`
	DurationSec int       `json:"duration_sec"`
	PriceCents  int       `json:"price_cents"`
	Status      string    `json:"status"` // available | booked | paused
	CreatedAt   time.Time `json:"created_at"`
}

type CreateRequest struct {
	DeviceID    string  `json:"device_id"`
	Label       *string `json:"label,omitempty"`
	DaysOfWeek  []int   `json:"days_of_week"`
	StartTime   string  `json:"start_time"`
	EndTime     string  `json:"end_time"`
	DurationSec int     `json:"duration_sec"`
	PriceCents  int     `json:"price_cents"`
}

type UpdateRequest struct {
	Label       *string `json:"label,omitempty"`
	DaysOfWeek  []int   `json:"days_of_week,omitempty"`
	StartTime   *string `json:"start_time,omitempty"`
	EndTime     *string `json:"end_time,omitempty"`
	DurationSec *int    `json:"duration_sec,omitempty"`
	PriceCents  *int    `json:"price_cents,omitempty"`
	Status      *string `json:"status,omitempty"`
}

type ListFilter struct {
	DeviceID string
	Status   string
}
