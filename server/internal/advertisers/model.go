package advertisers

import "time"

type Advertiser struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}

type CreateRequest struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}
