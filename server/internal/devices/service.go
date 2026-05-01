package devices

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Service struct {
	db *pgxpool.Pool
}

func NewService(db *pgxpool.Pool) *Service {
	return &Service{db: db}
}

func (s *Service) Register(ctx context.Context, req RegisterRequest) (RegisterResponse, error) {
	id := uuid.New().String()
	// Token is a random UUID for now; replace with a signed JWT in production.
	token := uuid.New().String()

	_, err := s.db.Exec(ctx, `
		INSERT INTO devices (id, venue_id, name, firmware_version, status)
		VALUES ($1, $2, $3, $4, 'offline')`,
		id, req.VenueID, req.Name, req.FirmwareVersion,
	)
	if err != nil {
		return RegisterResponse{}, fmt.Errorf("insert device: %w", err)
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

	return HeartbeatResponse{Status: "ok", Playlist: []any{}}, nil
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
