package scheduler

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type AdItem struct {
	BookingID   string `json:"booking_id"`
	CreativeURL string `json:"creative_url"`
	DurationSec int    `json:"duration_sec"`
}

type Service struct {
	db *pgxpool.Pool
}

func NewService(db *pgxpool.Pool) *Service {
	return &Service{db: db}
}

// PlaylistForDevice returns the ads currently scheduled for a device.
// It matches bookings whose slot covers the current day-of-week and time window.
func (s *Service) PlaylistForDevice(ctx context.Context, deviceID string) ([]AdItem, error) {
	now := time.Now().UTC()
	dow := int(now.Weekday()) // 0=Sun … 6=Sat
	clock := fmt.Sprintf("%02d:%02d", now.Hour(), now.Minute())

	rows, err := s.db.Query(ctx, `
		SELECT b.id, b.creative_url, s.duration_sec
		FROM bookings b
		JOIN ad_slots s ON s.id = b.slot_id
		WHERE s.device_id  = $1
		  AND b.status     IN ('pending', 'active')
		  AND b.starts_on  <= CURRENT_DATE
		  AND b.ends_on    >= CURRENT_DATE
		  AND $2            = ANY(s.days_of_week)
		  AND s.start_time <= $3::time
		  AND s.end_time   >  $3::time
		  AND b.creative_url IS NOT NULL
		ORDER BY b.created_at`,
		deviceID, dow, clock,
	)
	if err != nil {
		return nil, fmt.Errorf("playlist query: %w", err)
	}
	defer rows.Close()

	var items []AdItem
	for rows.Next() {
		var item AdItem
		if err := rows.Scan(&item.BookingID, &item.CreativeURL, &item.DurationSec); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}
