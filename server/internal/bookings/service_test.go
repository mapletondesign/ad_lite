package bookings_test

import (
	"context"
	"testing"

	"github.com/mapletondesign/ad_lite/internal/bookings"
	"github.com/mapletondesign/ad_lite/internal/testhelper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateBooking(t *testing.T) {
	pool := testhelper.DB(t)

	// setup inserts prerequisite rows and returns a slot + advertiser ID for use in the booking.
	setup := func(t *testing.T) (slotID, advertiserID string) {
		t.Helper()
		testhelper.TruncateAll(t, pool)
		ctx := context.Background()

		var venueID string
		require.NoError(t, pool.QueryRow(ctx,
			`INSERT INTO venues (name) VALUES ('Test Venue') RETURNING id`,
		).Scan(&venueID))

		var deviceID string
		require.NoError(t, pool.QueryRow(ctx,
			`INSERT INTO devices (venue_id, name, firmware_version) VALUES ($1, 'TV-1', '1.0') RETURNING id`,
			venueID,
		).Scan(&deviceID))

		require.NoError(t, pool.QueryRow(ctx, `
			INSERT INTO ad_slots (device_id, days_of_week, start_time, end_time, price_cents)
			VALUES ($1, '{1,2,3,4,5}', '09:00', '17:00', 5000)
			RETURNING id`, deviceID,
		).Scan(&slotID))

		require.NoError(t, pool.QueryRow(ctx,
			`INSERT INTO advertisers (name, email) VALUES ('Acme', 'acme@example.com') RETURNING id`,
		).Scan(&advertiserID))

		return slotID, advertiserID
	}

	t.Run("happy path", func(t *testing.T) {
		slotID, advertiserID := setup(t)
		svc := bookings.NewService(pool, "")

		b, err := svc.Create(context.Background(), bookings.CreateRequest{
			SlotID:       slotID,
			AdvertiserID: advertiserID,
			StartsOn:     "2026-06-01",
			EndsOn:       "2026-06-30",
		}, "")
		require.NoError(t, err)
		assert.NotEmpty(t, b.ID)
		assert.Equal(t, slotID, b.SlotID)
		assert.Equal(t, advertiserID, b.AdvertiserID)
		assert.Equal(t, 5000, b.PriceCents)
		assert.Equal(t, "pending", b.Status)
		assert.Equal(t, "2026-06-01", b.StartsOn)
		assert.Equal(t, "2026-06-30", b.EndsOn)

		var slotStatus string
		require.NoError(t, pool.QueryRow(context.Background(),
			`SELECT status FROM ad_slots WHERE id = $1`, slotID,
		).Scan(&slotStatus))
		assert.Equal(t, "booked", slotStatus)
	})

	t.Run("double booking on overlapping dates is rejected", func(t *testing.T) {
		slotID, advertiserID := setup(t)
		svc := bookings.NewService(pool, "")
		ctx := context.Background()

		_, err := svc.Create(ctx, bookings.CreateRequest{
			SlotID:       slotID,
			AdvertiserID: advertiserID,
			StartsOn:     "2026-06-01",
			EndsOn:       "2026-06-30",
		}, "")
		require.NoError(t, err)

		_, err = svc.Create(ctx, bookings.CreateRequest{
			SlotID:       slotID,
			AdvertiserID: advertiserID,
			StartsOn:     "2026-06-15",
			EndsOn:       "2026-07-15",
		}, "")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "not available")
	})

	t.Run("non-overlapping consecutive bookings on the same slot are allowed", func(t *testing.T) {
		slotID, advertiserID := setup(t)
		svc := bookings.NewService(pool, "")
		ctx := context.Background()

		_, err := svc.Create(ctx, bookings.CreateRequest{
			SlotID:       slotID,
			AdvertiserID: advertiserID,
			StartsOn:     "2026-06-01",
			EndsOn:       "2026-06-30",
		}, "")
		require.NoError(t, err)

		// Reset slot status manually so the second booking can proceed.
		_, err = pool.Exec(ctx, `UPDATE ad_slots SET status = 'available' WHERE id = $1`, slotID)
		require.NoError(t, err)

		_, err = svc.Create(ctx, bookings.CreateRequest{
			SlotID:       slotID,
			AdvertiserID: advertiserID,
			StartsOn:     "2026-07-01",
			EndsOn:       "2026-07-31",
		}, "")
		require.NoError(t, err)
	})

	t.Run("paused slot cannot be booked", func(t *testing.T) {
		slotID, advertiserID := setup(t)
		ctx := context.Background()

		_, err := pool.Exec(ctx, `UPDATE ad_slots SET status = 'paused' WHERE id = $1`, slotID)
		require.NoError(t, err)

		svc := bookings.NewService(pool, "")
		_, err = svc.Create(ctx, bookings.CreateRequest{
			SlotID:       slotID,
			AdvertiserID: advertiserID,
			StartsOn:     "2026-06-01",
			EndsOn:       "2026-06-30",
		}, "")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "paused")
	})

	t.Run("starts_on after ends_on is rejected", func(t *testing.T) {
		slotID, advertiserID := setup(t)
		svc := bookings.NewService(pool, "")

		_, err := svc.Create(context.Background(), bookings.CreateRequest{
			SlotID:       slotID,
			AdvertiserID: advertiserID,
			StartsOn:     "2026-06-30",
			EndsOn:       "2026-06-01",
		}, "")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "starts_on must be on or before ends_on")
	})

	t.Run("missing slot_id", func(t *testing.T) {
		_, advertiserID := setup(t)
		svc := bookings.NewService(pool, "")

		_, err := svc.Create(context.Background(), bookings.CreateRequest{
			AdvertiserID: advertiserID,
			StartsOn:     "2026-06-01",
			EndsOn:       "2026-06-30",
		}, "")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "slot_id is required")
	})

	t.Run("missing advertiser_id", func(t *testing.T) {
		slotID, _ := setup(t)
		svc := bookings.NewService(pool, "")

		_, err := svc.Create(context.Background(), bookings.CreateRequest{
			SlotID:   slotID,
			StartsOn: "2026-06-01",
			EndsOn:   "2026-06-30",
		}, "")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "advertiser_id is required")
	})
}

func TestListBookings(t *testing.T) {
	pool := testhelper.DB(t)
	testhelper.TruncateAll(t, pool)
	ctx := context.Background()

	var venueID string
	require.NoError(t, pool.QueryRow(ctx,
		`INSERT INTO venues (name) VALUES ('Test Venue') RETURNING id`,
	).Scan(&venueID))

	var deviceID string
	require.NoError(t, pool.QueryRow(ctx,
		`INSERT INTO devices (venue_id, name, firmware_version) VALUES ($1, 'TV-1', '1.0') RETURNING id`,
		venueID,
	).Scan(&deviceID))

	var slotID string
	require.NoError(t, pool.QueryRow(ctx, `
		INSERT INTO ad_slots (device_id, days_of_week, start_time, end_time, price_cents)
		VALUES ($1, '{1}', '09:00', '17:00', 1000) RETURNING id`, deviceID,
	).Scan(&slotID))

	var advID string
	require.NoError(t, pool.QueryRow(ctx,
		`INSERT INTO advertisers (name, email) VALUES ('Co', 'co@example.com') RETURNING id`,
	).Scan(&advID))

	svc := bookings.NewService(pool, "")
	_, err := svc.Create(ctx, bookings.CreateRequest{
		SlotID:       slotID,
		AdvertiserID: advID,
		StartsOn:     "2026-06-01",
		EndsOn:       "2026-06-15",
	}, "")
	require.NoError(t, err)

	t.Run("list all returns the booking", func(t *testing.T) {
		all, err := svc.List(ctx, bookings.ListFilter{})
		require.NoError(t, err)
		assert.Len(t, all, 1)
	})

	t.Run("filter by advertiser_id", func(t *testing.T) {
		byAdv, err := svc.List(ctx, bookings.ListFilter{AdvertiserID: advID})
		require.NoError(t, err)
		assert.Len(t, byAdv, 1)
	})

	t.Run("filter by unknown advertiser_id returns empty", func(t *testing.T) {
		none, err := svc.List(ctx, bookings.ListFilter{AdvertiserID: "00000000-0000-0000-0000-000000000000"})
		require.NoError(t, err)
		assert.Empty(t, none)
	})
}
