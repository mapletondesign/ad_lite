package devices

import (
	"context"
	"crypto/rsa"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mapletondesign/ad_lite/internal/auth"
	"github.com/mapletondesign/ad_lite/internal/scheduler"
)

type Service struct {
	db         *pgxpool.Pool
	scheduler  *scheduler.Service
	privateKey *rsa.PrivateKey
}

func NewService(db *pgxpool.Pool, sched *scheduler.Service, privateKey *rsa.PrivateKey) *Service {
	return &Service{db: db, scheduler: sched, privateKey: privateKey}
}

func (s *Service) Register(ctx context.Context, req RegisterRequest) (RegisterResponse, error) {
	id := uuid.New().String()

	_, err := s.db.Exec(ctx, `
		INSERT INTO devices (id, venue_id, name, firmware_version, status)
		VALUES ($1, $2, $3, $4, 'offline')`,
		id, req.VenueID, req.Name, req.FirmwareVersion,
	)
	if err != nil {
		return RegisterResponse{}, fmt.Errorf("insert device: %w", err)
	}

	token, err := auth.IssueDeviceToken(s.privateKey, id)
	if err != nil {
		return RegisterResponse{}, fmt.Errorf("issue device token: %w", err)
	}

	return RegisterResponse{DeviceID: id, Token: token}, nil
}

func (s *Service) Heartbeat(ctx context.Context, deviceID string, req HeartbeatRequest) (HeartbeatResponse, error) {
	now := time.Now().UTC()
	tag, err := s.db.Exec(ctx, `
		UPDATE devices
		SET status = 'online', last_seen = $1, ip_address = $2, firmware_version = $3, updated_at = $1
		WHERE id = $4`,
		now, req.IPAddress, req.FirmwareVersion, deviceID,
	)
	if err != nil {
		return HeartbeatResponse{}, fmt.Errorf("update device: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return HeartbeatResponse{}, fmt.Errorf("device not found")
	}

	playlist, err := s.scheduler.PlaylistForDevice(ctx, deviceID)
	if err != nil {
		// A scheduler failure should not take the device offline — return empty playlist.
		playlist = []scheduler.AdItem{}
	}

	return HeartbeatResponse{Status: "ok", Playlist: playlist}, nil
}

func (s *Service) RecordImpression(ctx context.Context, deviceID string, req ImpressionRequest) error {
	if req.BookingID == "" {
		return fmt.Errorf("booking_id is required")
	}
	_, err := s.db.Exec(ctx, `
		INSERT INTO impressions (booking_id, device_id)
		VALUES ($1, $2)`,
		req.BookingID, deviceID,
	)
	if err != nil {
		return fmt.Errorf("insert impression: %w", err)
	}
	return nil
}

func (s *Service) List(ctx context.Context) ([]Device, error) {
	rows, err := s.db.Query(ctx, `
		SELECT id, venue_id, name, status, last_seen, ip_address, firmware_version, created_at
		FROM devices ORDER BY created_at DESC`)
	if err != nil {
		return nil, fmt.Errorf("query devices: %w", err)
	}
	defer rows.Close()

	var devices []Device
	for rows.Next() {
		var d Device
		if err := rows.Scan(
			&d.ID, &d.VenueID, &d.Name, &d.Status,
			&d.LastSeen, &d.IPAddress, &d.FirmwareVersion, &d.CreatedAt,
		); err != nil {
			return nil, err
		}
		devices = append(devices, d)
	}
	return devices, rows.Err()
}
