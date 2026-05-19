package scheduler

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

const playlistTTL = 60 * time.Second

type AdItem struct {
	BookingID   string `json:"booking_id"`
	CreativeURL string `json:"creative_url"`
	DurationSec int    `json:"duration_sec"`
}

type Service struct {
	db  *pgxpool.Pool
	rdb *redis.Client
}

func NewService(db *pgxpool.Pool, rdb *redis.Client) *Service {
	return &Service{db: db, rdb: rdb}
}

func (s *Service) PlaylistForDevice(ctx context.Context, deviceID string) ([]AdItem, error) {
	key := "playlist:" + deviceID

	if cached, err := s.rdb.Get(ctx, key).Bytes(); err == nil {
		var items []AdItem
		if json.Unmarshal(cached, &items) == nil {
			return items, nil
		}
	}

	items, err := s.queryPlaylist(ctx, deviceID)
	if err != nil {
		return nil, err
	}

	if b, err := json.Marshal(items); err == nil {
		s.rdb.Set(ctx, key, b, playlistTTL)
	}

	return items, nil
}

func (s *Service) queryPlaylist(ctx context.Context, deviceID string) ([]AdItem, error) {
	now := time.Now().UTC()
	dow := int(now.Weekday())
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
