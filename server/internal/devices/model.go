package devices

import (
	"time"

	"github.com/mapletondesign/ad_pack/internal/scheduler"
)

type Device struct {
	ID              string     `json:"id"`
	VenueID         string     `json:"venue_id"`
	Name            string     `json:"name"`
	Status          string     `json:"status"`
	LastSeen        *time.Time `json:"last_seen,omitempty"`
	IPAddress       *string    `json:"ip_address,omitempty"`
	FirmwareVersion *string    `json:"firmware_version,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
}

type RegisterRequest struct {
	VenueID         string `json:"venue_id"`
	Name            string `json:"name"`
	FirmwareVersion string `json:"firmware_version"`
}

type RegisterResponse struct {
	DeviceID string `json:"device_id"`
	Token    string `json:"token"`
}

type HeartbeatRequest struct {
	IPAddress       string `json:"ip_address"`
	FirmwareVersion string `json:"firmware_version"`
}

type HeartbeatResponse struct {
	Status   string            `json:"status"`
	Playlist []scheduler.AdItem `json:"playlist"`
}
