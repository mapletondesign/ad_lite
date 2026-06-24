package devices_test

import (
	"context"
	"testing"
	"time"

	"github.com/mapletondesign/ad_lite/internal/auth"
	"github.com/mapletondesign/ad_lite/internal/devices"
	"github.com/mapletondesign/ad_lite/internal/scheduler"
	"github.com/mapletondesign/ad_lite/internal/testhelper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDeviceRegister(t *testing.T) {
	pool := testhelper.DB(t)
	priv, pub := testhelper.RSAKeys(t)
	rdb := testhelper.Redis(t)
	testhelper.TruncateAll(t, pool)
	ctx := context.Background()

	var venueID string
	require.NoError(t, pool.QueryRow(ctx,
		`INSERT INTO venues (name) VALUES ('Reg Venue') RETURNING id`,
	).Scan(&venueID))

	svc := devices.NewService(pool, scheduler.NewService(pool, rdb), priv)

	t.Run("creates device row and returns a signed device JWT", func(t *testing.T) {
		resp, err := svc.Register(ctx, devices.RegisterRequest{
			VenueID:         venueID,
			Name:            "Lobby Screen",
			FirmwareVersion: "1.0.0",
		})
		require.NoError(t, err)
		assert.NotEmpty(t, resp.DeviceID)
		assert.NotEmpty(t, resp.Token)

		// Token must parse and carry the device role.
		claims, err := auth.VerifyToken(pub, resp.Token)
		require.NoError(t, err)
		assert.Equal(t, "device", claims.Role)
		assert.Equal(t, resp.DeviceID, claims.Subject)

		// Device row must exist in the database.
		var name string
		require.NoError(t, pool.QueryRow(ctx,
			`SELECT name FROM devices WHERE id = $1`, resp.DeviceID,
		).Scan(&name))
		assert.Equal(t, "Lobby Screen", name)
	})
}

func TestDeviceHeartbeat(t *testing.T) {
	pool := testhelper.DB(t)
	priv, _ := testhelper.RSAKeys(t)
	rdb := testhelper.Redis(t)
	testhelper.TruncateAll(t, pool)
	ctx := context.Background()

	var venueID string
	require.NoError(t, pool.QueryRow(ctx,
		`INSERT INTO venues (name) VALUES ('HB Venue') RETURNING id`,
	).Scan(&venueID))

	svc := devices.NewService(pool, scheduler.NewService(pool, rdb), priv)

	resp, err := svc.Register(ctx, devices.RegisterRequest{
		VenueID:         venueID,
		Name:            "HB Screen",
		FirmwareVersion: "1.0.0",
	})
	require.NoError(t, err)
	deviceID := resp.DeviceID

	t.Run("heartbeat sets status to online and updates last_seen", func(t *testing.T) {
		before := time.Now().UTC().Add(-time.Second)

		hbResp, err := svc.Heartbeat(ctx, deviceID, devices.HeartbeatRequest{
			IPAddress:       "192.168.1.10",
			FirmwareVersion: "1.0.1",
		})
		require.NoError(t, err)
		assert.Equal(t, "ok", hbResp.Status)
		// Playlist may be nil or empty when no active bookings exist — both are valid.
		assert.Empty(t, hbResp.Playlist)

		var status string
		var lastSeen time.Time
		require.NoError(t, pool.QueryRow(ctx,
			`SELECT status, last_seen FROM devices WHERE id = $1`, deviceID,
		).Scan(&status, &lastSeen))
		assert.Equal(t, "online", status)
		assert.True(t, lastSeen.After(before), "last_seen should be updated")
	})

	t.Run("heartbeat for unknown device returns error", func(t *testing.T) {
		_, err := svc.Heartbeat(ctx, "00000000-0000-0000-0000-000000000000", devices.HeartbeatRequest{
			IPAddress:       "10.0.0.1",
			FirmwareVersion: "1.0.0",
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "device not found")
	})
}

func TestRecordImpression(t *testing.T) {
	pool := testhelper.DB(t)
	priv, _ := testhelper.RSAKeys(t)
	rdb := testhelper.Redis(t)
	testhelper.TruncateAll(t, pool)
	ctx := context.Background()

	// Set up venue → device → slot → advertiser → booking.
	var venueID string
	require.NoError(t, pool.QueryRow(ctx,
		`INSERT INTO venues (name) VALUES ('Imp Venue') RETURNING id`,
	).Scan(&venueID))

	var deviceID string
	require.NoError(t, pool.QueryRow(ctx,
		`INSERT INTO devices (venue_id, name, firmware_version) VALUES ($1, 'Imp Screen', '1.0') RETURNING id`,
		venueID,
	).Scan(&deviceID))

	var slotID string
	require.NoError(t, pool.QueryRow(ctx, `
		INSERT INTO ad_slots (device_id, days_of_week, start_time, end_time, price_cents)
		VALUES ($1, '{1}', '09:00', '17:00', 500) RETURNING id`, deviceID,
	).Scan(&slotID))

	var advID string
	require.NoError(t, pool.QueryRow(ctx,
		`INSERT INTO advertisers (name, email) VALUES ('ImpCo', 'imp@example.com') RETURNING id`,
	).Scan(&advID))

	var bookingID string
	require.NoError(t, pool.QueryRow(ctx, `
		INSERT INTO bookings (slot_id, advertiser_id, starts_on, ends_on, price_cents, status)
		VALUES ($1, $2, CURRENT_DATE, CURRENT_DATE + 30, 500, 'active') RETURNING id`,
		slotID, advID,
	).Scan(&bookingID))

	svc := devices.NewService(pool, scheduler.NewService(pool, rdb), priv)

	t.Run("impression is recorded in the database", func(t *testing.T) {
		err := svc.RecordImpression(ctx, deviceID, devices.ImpressionRequest{BookingID: bookingID})
		require.NoError(t, err)

		var count int
		require.NoError(t, pool.QueryRow(ctx,
			`SELECT COUNT(*) FROM impressions WHERE booking_id = $1 AND device_id = $2`,
			bookingID, deviceID,
		).Scan(&count))
		assert.Equal(t, 1, count)
	})

	t.Run("impression without booking_id is rejected", func(t *testing.T) {
		err := svc.RecordImpression(ctx, deviceID, devices.ImpressionRequest{})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "booking_id is required")
	})
}
