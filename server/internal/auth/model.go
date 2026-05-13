package auth

import "time"

type User struct {
	ID           string
	Email        string
	PasswordHash string
	Role         string
	AdvertiserID *string
	VenueID      *string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Role     string `json:"role"` // "advertiser" | "venue"
	Name     string `json:"name"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"` // seconds
}
