package slots

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

func (s *Service) Create(ctx context.Context, req CreateRequest) (Slot, error) {
	if err := validateCreate(req); err != nil {
		return Slot{}, err
	}
	if req.DurationSec == 0 {
		req.DurationSec = 15
	}

	id := uuid.New().String()
	var slot Slot
	err := s.db.QueryRow(ctx, `
		INSERT INTO ad_slots (id, device_id, label, days_of_week, start_time, end_time, duration_sec, price_cents)
		VALUES ($1, $2, $3, $4, $5::time, $6::time, $7, $8)
		RETURNING id, device_id, label, days_of_week,
		          to_char(start_time, 'HH24:MI'), to_char(end_time, 'HH24:MI'),
		          duration_sec, price_cents, status, created_at`,
		id, req.DeviceID, req.Label, req.DaysOfWeek,
		req.StartTime, req.EndTime, req.DurationSec, req.PriceCents,
	).Scan(
		&slot.ID, &slot.DeviceID, &slot.Label, &slot.DaysOfWeek,
		&slot.StartTime, &slot.EndTime,
		&slot.DurationSec, &slot.PriceCents, &slot.Status, &slot.CreatedAt,
	)
	if err != nil {
		return Slot{}, fmt.Errorf("insert slot: %w", err)
	}
	return slot, nil
}

func (s *Service) List(ctx context.Context, f ListFilter) ([]Slot, error) {
	query := `
		SELECT id, device_id, label, days_of_week,
		       to_char(start_time, 'HH24:MI'), to_char(end_time, 'HH24:MI'),
		       duration_sec, price_cents, status, created_at
		FROM ad_slots
		WHERE ($1 = '' OR device_id::text = $1)
		  AND ($2 = '' OR status         = $2)
		ORDER BY created_at DESC`

	rows, err := s.db.Query(ctx, query, f.DeviceID, f.Status)
	if err != nil {
		return nil, fmt.Errorf("query slots: %w", err)
	}
	defer rows.Close()

	var out []Slot
	for rows.Next() {
		var slot Slot
		if err := rows.Scan(
			&slot.ID, &slot.DeviceID, &slot.Label, &slot.DaysOfWeek,
			&slot.StartTime, &slot.EndTime,
			&slot.DurationSec, &slot.PriceCents, &slot.Status, &slot.CreatedAt,
		); err != nil {
			return nil, err
		}
		out = append(out, slot)
	}
	return out, rows.Err()
}

func (s *Service) Update(ctx context.Context, id string, req UpdateRequest) (Slot, error) {
	// Fetch current, apply only provided fields, write back.
	var cur Slot
	err := s.db.QueryRow(ctx, `
		SELECT id, device_id, label, days_of_week,
		       to_char(start_time, 'HH24:MI'), to_char(end_time, 'HH24:MI'),
		       duration_sec, price_cents, status, created_at
		FROM ad_slots WHERE id = $1`, id,
	).Scan(
		&cur.ID, &cur.DeviceID, &cur.Label, &cur.DaysOfWeek,
		&cur.StartTime, &cur.EndTime,
		&cur.DurationSec, &cur.PriceCents, &cur.Status, &cur.CreatedAt,
	)
	if err != nil {
		return Slot{}, fmt.Errorf("slot not found")
	}

	if req.Label != nil {
		cur.Label = req.Label
	}
	if len(req.DaysOfWeek) > 0 {
		cur.DaysOfWeek = req.DaysOfWeek
	}
	if req.StartTime != nil {
		cur.StartTime = *req.StartTime
	}
	if req.EndTime != nil {
		cur.EndTime = *req.EndTime
	}
	if req.DurationSec != nil {
		cur.DurationSec = *req.DurationSec
	}
	if req.PriceCents != nil {
		cur.PriceCents = *req.PriceCents
	}
	if req.Status != nil {
		if err := validateStatus(*req.Status); err != nil {
			return Slot{}, err
		}
		cur.Status = *req.Status
	}

	var updated Slot
	err = s.db.QueryRow(ctx, `
		UPDATE ad_slots
		SET label=$2, days_of_week=$3, start_time=$4::time, end_time=$5::time,
		    duration_sec=$6, price_cents=$7, status=$8
		WHERE id=$1
		RETURNING id, device_id, label, days_of_week,
		          to_char(start_time, 'HH24:MI'), to_char(end_time, 'HH24:MI'),
		          duration_sec, price_cents, status, created_at`,
		cur.ID, cur.Label, cur.DaysOfWeek, cur.StartTime, cur.EndTime,
		cur.DurationSec, cur.PriceCents, cur.Status,
	).Scan(
		&updated.ID, &updated.DeviceID, &updated.Label, &updated.DaysOfWeek,
		&updated.StartTime, &updated.EndTime,
		&updated.DurationSec, &updated.PriceCents, &updated.Status, &updated.CreatedAt,
	)
	if err != nil {
		return Slot{}, fmt.Errorf("update slot: %w", err)
	}
	return updated, nil
}

func validateCreate(req CreateRequest) error {
	if req.DeviceID == "" {
		return fmt.Errorf("device_id is required")
	}
	if len(req.DaysOfWeek) == 0 {
		return fmt.Errorf("days_of_week must not be empty")
	}
	if req.StartTime == "" || req.EndTime == "" {
		return fmt.Errorf("start_time and end_time are required")
	}
	if req.PriceCents <= 0 {
		return fmt.Errorf("price_cents must be greater than zero")
	}
	return nil
}

func (s *Service) Delete(ctx context.Context, id string) error {
	var count int
	err := s.db.QueryRow(ctx, `
		SELECT COUNT(*) FROM bookings
		WHERE slot_id = $1 AND status IN ('pending', 'active')`, id,
	).Scan(&count)
	if err != nil {
		return fmt.Errorf("check bookings: %w", err)
	}
	if count > 0 {
		return fmt.Errorf("slot has active or pending bookings and cannot be deleted")
	}

	tag, err := s.db.Exec(ctx, `DELETE FROM ad_slots WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete slot: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("slot not found")
	}
	return nil
}

func validateStatus(s string) error {
	switch s {
	case "available", "booked", "paused":
		return nil
	}
	return fmt.Errorf("status must be available, booked, or paused")
}
