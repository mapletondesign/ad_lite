package bookings

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Service struct {
	db *pgxpool.Pool
}

func NewService(db *pgxpool.Pool) *Service {
	return &Service{db: db}
}

func (s *Service) Create(ctx context.Context, req CreateRequest) (Booking, error) {
	if err := validateCreate(req); err != nil {
		return Booking{}, err
	}

	// Use a transaction so the availability check and insert are atomic.
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return Booking{}, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	// Check for overlapping active/pending bookings on this slot.
	var conflicts int
	err = tx.QueryRow(ctx, `
		SELECT COUNT(*) FROM bookings
		WHERE slot_id = $1
		  AND status IN ('pending', 'active')
		  AND starts_on <= $3::date
		  AND ends_on   >= $2::date`,
		req.SlotID, req.StartsOn, req.EndsOn,
	).Scan(&conflicts)
	if err != nil {
		return Booking{}, fmt.Errorf("availability check: %w", err)
	}
	if conflicts > 0 {
		return Booking{}, fmt.Errorf("slot is not available for the requested date range")
	}

	// Fetch slot price to set booking price.
	var priceCents int
	var slotStatus string
	err = tx.QueryRow(ctx,
		`SELECT price_cents, status FROM ad_slots WHERE id = $1`,
		req.SlotID,
	).Scan(&priceCents, &slotStatus)
	if err != nil {
		return Booking{}, fmt.Errorf("slot not found")
	}
	if slotStatus == "paused" {
		return Booking{}, fmt.Errorf("slot is paused and cannot be booked")
	}

	id := uuid.New().String()
	var b Booking
	err = tx.QueryRow(ctx, `
		INSERT INTO bookings (id, slot_id, advertiser_id, creative_url, starts_on, ends_on, price_cents)
		VALUES ($1, $2, $3, $4, $5::date, $6::date, $7)
		RETURNING id, slot_id, advertiser_id, creative_url,
		          to_char(starts_on, 'YYYY-MM-DD'), to_char(ends_on, 'YYYY-MM-DD'),
		          price_cents, status, created_at`,
		id, req.SlotID, req.AdvertiserID, req.CreativeURL,
		req.StartsOn, req.EndsOn, priceCents,
	).Scan(
		&b.ID, &b.SlotID, &b.AdvertiserID, &b.CreativeURL,
		&b.StartsOn, &b.EndsOn,
		&b.PriceCents, &b.Status, &b.CreatedAt,
	)
	if err != nil {
		return Booking{}, fmt.Errorf("insert booking: %w", err)
	}

	// Mark the slot as booked.
	_, err = tx.Exec(ctx,
		`UPDATE ad_slots SET status = 'booked' WHERE id = $1`,
		req.SlotID,
	)
	if err != nil {
		return Booking{}, fmt.Errorf("update slot status: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return Booking{}, fmt.Errorf("commit: %w", err)
	}
	return b, nil
}

func (s *Service) List(ctx context.Context, f ListFilter) ([]Booking, error) {
	rows, err := s.db.Query(ctx, `
		SELECT id, slot_id, advertiser_id, creative_url,
		       to_char(starts_on, 'YYYY-MM-DD'), to_char(ends_on, 'YYYY-MM-DD'),
		       price_cents, status, created_at
		FROM bookings
		WHERE ($1 = '' OR slot_id       = $1)
		  AND ($2 = '' OR advertiser_id = $2)
		  AND ($3 = '' OR status        = $3)
		ORDER BY created_at DESC`,
		f.SlotID, f.AdvertiserID, f.Status,
	)
	if err != nil {
		return nil, fmt.Errorf("query bookings: %w", err)
	}
	defer rows.Close()

	var out []Booking
	for rows.Next() {
		var b Booking
		if err := rows.Scan(
			&b.ID, &b.SlotID, &b.AdvertiserID, &b.CreativeURL,
			&b.StartsOn, &b.EndsOn,
			&b.PriceCents, &b.Status, &b.CreatedAt,
		); err != nil {
			return nil, err
		}
		out = append(out, b)
	}
	return out, rows.Err()
}

func validateCreate(req CreateRequest) error {
	if req.SlotID == "" {
		return fmt.Errorf("slot_id is required")
	}
	if req.AdvertiserID == "" {
		return fmt.Errorf("advertiser_id is required")
	}
	if req.StartsOn == "" || req.EndsOn == "" {
		return fmt.Errorf("starts_on and ends_on are required")
	}
	if req.StartsOn > req.EndsOn {
		return fmt.Errorf("starts_on must be on or before ends_on")
	}
	return nil
}
