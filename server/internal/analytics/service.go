package analytics

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Service struct {
	db *pgxpool.Pool
}

func NewService(db *pgxpool.Pool) *Service {
	return &Service{db: db}
}

type ImpressionRow struct {
	BookingID    string `json:"booking_id"`
	AdvertiserID string `json:"advertiser_id"`
	DeviceID     string `json:"device_id"`
	Date         string `json:"date"` // "YYYY-MM-DD"
	Count        int    `json:"count"`
}

type ImpressionFilter struct {
	AdvertiserID string // restrict to one advertiser's bookings
	BookingID    string
	DeviceID     string
	From         string // "YYYY-MM-DD"; empty = no lower bound
	To           string // "YYYY-MM-DD"; empty = no upper bound
}

func (s *Service) Impressions(ctx context.Context, f ImpressionFilter) ([]ImpressionRow, error) {
	rows, err := s.db.Query(ctx, `
		SELECT i.booking_id::text, b.advertiser_id::text, i.device_id::text,
		       to_char(i.played_at::date, 'YYYY-MM-DD'), COUNT(*)
		FROM impressions i
		JOIN bookings b ON b.id = i.booking_id
		WHERE ($1 = '' OR b.advertiser_id::text = $1)
		  AND ($2 = '' OR i.booking_id::text   = $2)
		  AND ($3 = '' OR i.device_id::text    = $3)
		  AND ($4 = '' OR i.played_at >= $4::timestamptz)
		  AND ($5 = '' OR i.played_at <  ($5::date + interval '1 day')::timestamptz)
		GROUP BY i.booking_id, b.advertiser_id, i.device_id, i.played_at::date
		ORDER BY i.played_at::date DESC, i.booking_id`,
		f.AdvertiserID, f.BookingID, f.DeviceID, f.From, f.To,
	)
	if err != nil {
		return nil, fmt.Errorf("query impressions: %w", err)
	}
	defer rows.Close()

	var out []ImpressionRow
	for rows.Next() {
		var r ImpressionRow
		if err := rows.Scan(&r.BookingID, &r.AdvertiserID, &r.DeviceID, &r.Date, &r.Count); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	if out == nil {
		out = []ImpressionRow{}
	}
	return out, rows.Err()
}
