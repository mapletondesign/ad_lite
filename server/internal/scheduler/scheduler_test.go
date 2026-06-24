package scheduler_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/mapletondesign/ad_lite/internal/scheduler"
	"github.com/mapletondesign/ad_lite/internal/testhelper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPlaylistForDevice(t *testing.T) {
	pool := testhelper.DB(t)
	rdb := testhelper.Redis(t)

	now := time.Now().UTC()
	today := now.Format("2006-01-02")
	dow := int(now.Weekday())

	// Window: start 2 minutes ago, end 2 hours from now — guaranteed to include current time.
	windowStart := fmt.Sprintf("%02d:%02d", now.Add(-2*time.Minute).Hour(), now.Add(-2*time.Minute).Minute())
	windowEnd := fmt.Sprintf("%02d:%02d", now.Add(2*time.Hour).Hour(), now.Add(2*time.Hour).Minute())

	setup := func(t *testing.T) (deviceID, bookingID string) {
		t.Helper()
		testhelper.TruncateAll(t, pool)
		ctx := context.Background()

		var venueID string
		require.NoError(t, pool.QueryRow(ctx,
			`INSERT INTO venues (name) VALUES ('Sched Venue') RETURNING id`,
		).Scan(&venueID))

		require.NoError(t, pool.QueryRow(ctx,
			`INSERT INTO devices (venue_id, name, firmware_version) VALUES ($1, 'TV-Sched', '1.0') RETURNING id`,
			venueID,
		).Scan(&deviceID))

		var slotID string
		require.NoError(t, pool.QueryRow(ctx, `
			INSERT INTO ad_slots (device_id, days_of_week, start_time, end_time, duration_sec, price_cents)
			VALUES ($1, $2, $3, $4, 15, 1000) RETURNING id`,
			deviceID, []int{dow}, windowStart, windowEnd,
		).Scan(&slotID))

		var advID string
		require.NoError(t, pool.QueryRow(ctx,
			`INSERT INTO advertisers (name, email) VALUES ('SchedCo', 'sched@example.com') RETURNING id`,
		).Scan(&advID))

		creativeURL := "https://cdn.example.com/ad.mp4"
		require.NoError(t, pool.QueryRow(ctx, `
			INSERT INTO bookings (slot_id, advertiser_id, creative_url, starts_on, ends_on, price_cents, status)
			VALUES ($1, $2, $3, $4::date, $4::date, 1000, 'active') RETURNING id`,
			slotID, advID, creativeURL, today,
		).Scan(&bookingID))

		return deviceID, bookingID
	}

	t.Run("active booking in current time window appears in playlist", func(t *testing.T) {
		deviceID, bookingID := setup(t)
		svc := scheduler.NewService(pool, rdb)

		items, err := svc.PlaylistForDevice(context.Background(), deviceID)
		require.NoError(t, err)
		require.Len(t, items, 1)
		assert.Equal(t, bookingID, items[0].BookingID)
		assert.Equal(t, "https://cdn.example.com/ad.mp4", items[0].CreativeURL)
		assert.Equal(t, 15, items[0].DurationSec)
	})

	t.Run("second call returns cached result from Redis", func(t *testing.T) {
		deviceID, _ := setup(t)
		svc := scheduler.NewService(pool, rdb)
		ctx := context.Background()

		first, err := svc.PlaylistForDevice(ctx, deviceID)
		require.NoError(t, err)

		// Corrupt the DB data — the cache should shield us from seeing the change.
		_, err = pool.Exec(ctx, `UPDATE bookings SET creative_url = 'corrupted'`)
		require.NoError(t, err)

		second, err := svc.PlaylistForDevice(ctx, deviceID)
		require.NoError(t, err)
		assert.Equal(t, first, second, "expected cached result, not re-queried DB")
	})

	t.Run("booking outside date range is excluded", func(t *testing.T) {
		deviceID, bookingID := setup(t)
		ctx := context.Background()

		// Move the booking to yesterday—tomorrow so it ends before today.
		yesterday := now.AddDate(0, 0, -2).Format("2006-01-02")
		dayBefore := now.AddDate(0, 0, -1).Format("2006-01-02")
		_, err := pool.Exec(ctx,
			`UPDATE bookings SET starts_on = $1::date, ends_on = $2::date WHERE id = $3`,
			yesterday, dayBefore, bookingID,
		)
		require.NoError(t, err)

		svc := scheduler.NewService(pool, rdb)
		items, err := svc.PlaylistForDevice(ctx, deviceID)
		require.NoError(t, err)
		assert.Empty(t, items)
	})

	t.Run("booking on wrong day of week is excluded", func(t *testing.T) {
		deviceID, _ := setup(t)
		ctx := context.Background()

		// Set the slot to only run on a different day.
		wrongDOW := (dow + 1) % 7
		_, err := pool.Exec(ctx, `UPDATE ad_slots SET days_of_week = $1`, []int{wrongDOW})
		require.NoError(t, err)

		svc := scheduler.NewService(pool, rdb)
		items, err := svc.PlaylistForDevice(ctx, deviceID)
		require.NoError(t, err)
		assert.Empty(t, items)
	})

	t.Run("booking with no creative_url is excluded", func(t *testing.T) {
		deviceID, bookingID := setup(t)
		ctx := context.Background()

		_, err := pool.Exec(ctx, `UPDATE bookings SET creative_url = NULL WHERE id = $1`, bookingID)
		require.NoError(t, err)

		svc := scheduler.NewService(pool, rdb)
		items, err := svc.PlaylistForDevice(ctx, deviceID)
		require.NoError(t, err)
		assert.Empty(t, items)
	})

	t.Run("unknown device returns empty playlist", func(t *testing.T) {
		testhelper.TruncateAll(t, pool)
		svc := scheduler.NewService(pool, rdb)

		items, err := svc.PlaylistForDevice(context.Background(), "00000000-0000-0000-0000-000000000000")
		require.NoError(t, err)
		assert.Empty(t, items)
	})
}
