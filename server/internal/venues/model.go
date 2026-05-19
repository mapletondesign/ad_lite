package venues

import "time"

type Venue struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Category  *string   `json:"category,omitempty"`
	Address   *string   `json:"address,omitempty"`
	City      *string   `json:"city,omitempty"`
	State     *string   `json:"state,omitempty"`
	Country   string    `json:"country"`
	CreatedAt time.Time `json:"created_at"`
}

type CreateRequest struct {
	Name     string `json:"name"`
	Category string `json:"category"`
	Address  string `json:"address"`
	City     string `json:"city"`
	State    string `json:"state"`
}
