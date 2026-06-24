package bookings

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Service struct {
	db      *pgxpool.Pool
	cdnBase string // required URL prefix for creative_url; empty = no restriction
}

func NewService(db *pgxpool.Pool, cdnBase string) *Service {
	return &Service{db: db, cdnBase: cdnBase}
}

func (s *Service) Create(ctx context.Context, req CreateRequest, actorID string) (Booking, error) {
	if err := validateCreate(req); err != nil {
		return Booking{}, err
	}
	if err := s.validateCreativeURL(req.CreativeURL); err != nil {
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

	var createdBy *string
	if actorID != "" {
		createdBy = &actorID
	}

	id := uuid.New().String()
	var b Booking
	err = tx.QueryRow(ctx, `
		INSERT INTO bookings (id, slot_id, advertiser_id, creative_url, starts_on, ends_on, price_cents, created_by)
		VALUES ($1, $2, $3, $4, $5::date, $6::date, $7, $8)
		RETURNING id, slot_id, advertiser_id, creative_url,
		          to_char(starts_on, 'YYYY-MM-DD'), to_char(ends_on, 'YYYY-MM-DD'),
		          price_cents, status, created_at`,
		id, req.SlotID, req.AdvertiserID, req.CreativeURL,
		req.StartsOn, req.EndsOn, priceCents, createdBy,
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
		WHERE ($1 = '' OR slot_id::text       = $1)
		  AND ($2 = '' OR advertiser_id::text = $2)
		  AND ($3 = '' OR status              = $3)
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

// AdminUpdate allows an admin to change the status and/or creative_url of any booking.
func (s *Service) AdminUpdate(ctx context.Context, id string, req AdminUpdateRequest, actorID string) (Booking, error) {
	if req.Status != nil {
		switch *req.Status {
		case "pending", "active", "completed", "cancelled":
		default:
			return Booking{}, fmt.Errorf("invalid status")
		}
	}
	if err := s.validateCreativeURL(req.CreativeURL); err != nil {
		return Booking{}, err
	}

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return Booking{}, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	var updatedBy *string
	if actorID != "" {
		updatedBy = &actorID
	}

	var b Booking
	err = tx.QueryRow(ctx, `
		UPDATE bookings
		SET status       = COALESCE($2, status),
		    creative_url = COALESCE($3, creative_url),
		    updated_by   = $4,
		    updated_at   = now()
		WHERE id = $1
		RETURNING id, slot_id, advertiser_id, creative_url,
		          to_char(starts_on, 'YYYY-MM-DD'), to_char(ends_on, 'YYYY-MM-DD'),
		          price_cents, status, created_at`,
		id, req.Status, req.CreativeURL, updatedBy,
	).Scan(&b.ID, &b.SlotID, &b.AdvertiserID, &b.CreativeURL,
		&b.StartsOn, &b.EndsOn, &b.PriceCents, &b.Status, &b.CreatedAt)
	if err != nil {
		return Booking{}, fmt.Errorf("booking not found")
	}

	if req.Status != nil && *req.Status == "cancelled" {
		if err := releaseSlot(ctx, tx, b.SlotID); err != nil {
			return Booking{}, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return Booking{}, fmt.Errorf("commit: %w", err)
	}
	return b, nil
}

// AdvertiserUpdate lets an advertiser cancel their own booking or update creative_url.
func (s *Service) AdvertiserUpdate(ctx context.Context, id, advertiserID string, req AdvertiserUpdateRequest, actorID string) (Booking, error) {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return Booking{}, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	var b Booking
	err = tx.QueryRow(ctx, `
		SELECT id, slot_id, advertiser_id, creative_url,
		       to_char(starts_on, 'YYYY-MM-DD'), to_char(ends_on, 'YYYY-MM-DD'),
		       price_cents, status, created_at
		FROM bookings WHERE id = $1`, id,
	).Scan(&b.ID, &b.SlotID, &b.AdvertiserID, &b.CreativeURL,
		&b.StartsOn, &b.EndsOn, &b.PriceCents, &b.Status, &b.CreatedAt)
	if err != nil {
		return Booking{}, fmt.Errorf("booking not found")
	}
	if b.AdvertiserID != advertiserID {
		return Booking{}, fmt.Errorf("forbidden")
	}

	if req.Cancel {
		if b.Status == "completed" || b.Status == "cancelled" {
			return Booking{}, fmt.Errorf("booking cannot be cancelled in its current state")
		}
		b.Status = "cancelled"
	}
	if req.CreativeURL != nil {
		if err := s.validateCreativeURL(req.CreativeURL); err != nil {
			return Booking{}, err
		}
		b.CreativeURL = req.CreativeURL
	}

	var updatedBy *string
	if actorID != "" {
		updatedBy = &actorID
	}

	err = tx.QueryRow(ctx, `
		UPDATE bookings
		SET status       = $2,
		    creative_url = $3,
		    updated_by   = $4,
		    updated_at   = now()
		WHERE id = $1
		RETURNING id, slot_id, advertiser_id, creative_url,
		          to_char(starts_on, 'YYYY-MM-DD'), to_char(ends_on, 'YYYY-MM-DD'),
		          price_cents, status, created_at`,
		b.ID, b.Status, b.CreativeURL, updatedBy,
	).Scan(&b.ID, &b.SlotID, &b.AdvertiserID, &b.CreativeURL,
		&b.StartsOn, &b.EndsOn, &b.PriceCents, &b.Status, &b.CreatedAt)
	if err != nil {
		return Booking{}, fmt.Errorf("update booking: %w", err)
	}

	if req.Cancel {
		if err := releaseSlot(ctx, tx, b.SlotID); err != nil {
			return Booking{}, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return Booking{}, fmt.Errorf("commit: %w", err)
	}
	return b, nil
}

// releaseSlot sets the slot back to 'available' if no pending/active bookings remain.
func releaseSlot(ctx context.Context, tx pgx.Tx, slotID string) error {
	var remaining int
	err := tx.QueryRow(ctx, `
		SELECT COUNT(*) FROM bookings
		WHERE slot_id = $1 AND status IN ('pending', 'active')`, slotID,
	).Scan(&remaining)
	if err != nil {
		return fmt.Errorf("check remaining bookings: %w", err)
	}
	if remaining == 0 {
		if _, err := tx.Exec(ctx,
			`UPDATE ad_slots SET status = 'available' WHERE id = $1`, slotID,
		); err != nil {
			return fmt.Errorf("release slot: %w", err)
		}
	}
	return nil
}

func (s *Service) validateCreativeURL(u *string) error {
	if u == nil || *u == "" || s.cdnBase == "" {
		return nil
	}
	if !strings.HasPrefix(*u, s.cdnBase) {
		return fmt.Errorf("creative_url must be hosted on the CDN (%s)", s.cdnBase)
	}
	return nil
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
