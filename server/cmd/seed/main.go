package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/mapletondesign/ad_lite/internal/auth"
	"github.com/mapletondesign/ad_lite/internal/db"
)

func main() {
	email := flag.String("email", "admin@adlite.com", "admin user email")
	password := flag.String("password", "changeme", "admin user password")
	demo := flag.Bool("demo", false, "also seed demo venues, advertisers, devices, slots, and bookings")
	flag.Parse()

	if err := godotenv.Load(); err != nil {
		log.Println("no .env file, reading environment directly")
	}

	pool, err := db.Connect(os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatalf("db connect: %v", err)
	}
	defer pool.Close()

	ctx := context.Background()

	adminID, created, err := upsertAdmin(ctx, pool, *email, *password)
	if err != nil {
		log.Fatalf("seed admin: %v", err)
	}
	if created {
		fmt.Printf("✓ admin created  email=%s  id=%s\n", *email, adminID)
	} else {
		fmt.Printf("✓ admin exists   email=%s  id=%s\n", *email, adminID)
	}

	if *demo {
		if err := seedDemo(ctx, pool); err != nil {
			log.Fatalf("seed demo data: %v", err)
		}
	}
}

func upsertAdmin(ctx context.Context, pool *pgxpool.Pool, email, password string) (id string, created bool, err error) {
	err = pool.QueryRow(ctx, `SELECT id FROM users WHERE email = $1`, email).Scan(&id)
	if err == nil {
		return id, false, nil
	}

	hash, err := auth.HashPassword(password)
	if err != nil {
		return "", false, fmt.Errorf("hash password: %w", err)
	}

	err = pool.QueryRow(ctx,
		`INSERT INTO users (email, password_hash, role) VALUES ($1, $2, 'admin') RETURNING id`,
		email, hash,
	).Scan(&id)
	if err != nil {
		return "", false, fmt.Errorf("insert user: %w", err)
	}
	return id, true, nil
}

type venueRow struct {
	name, category, address, city, state string
}

type advertiserRow struct {
	name, email string
}

type deviceRow struct {
	name, status, ip, firmware string
	venueIdx                   int
}

type slotRow struct {
	label                    string
	daysOfWeek               []int
	startTime, endTime       string
	durationSec, priceCents  int
	status                   string
	deviceIdx                int
}

type bookingRow struct {
	slotIdx, advertiserIdx     int
	startsOn, endsOn           string
	priceCents                 int
	status                     string
}

func seedDemo(ctx context.Context, pool *pgxpool.Pool) error {
	venues := []venueRow{
		{"The Grind Coffee Co.", "restaurant", "840 W Randolph St", "Chicago", "IL"},
		{"FitZone Gym", "gym", "1200 N Milwaukee Ave", "Chicago", "IL"},
		{"River North Food Hall", "restaurant", "350 W Hubbard St", "Chicago", "IL"},
		{"Westfield Mall", "retail", "835 N Michigan Ave", "Chicago", "IL"},
	}

	advertisers := []advertiserRow{
		{"Bloom Florist", "hello@bloomflorist.com"},
		{"TechFast Insurance", "ads@techfast.com"},
		{"Velocity Energy Drink", "marketing@velocitydrink.com"},
		{"Cornerstone Realty", "media@cornerstonerealty.com"},
	}

	devices := []deviceRow{
		{"Counter Display", "online", "192.168.1.101", "1.4.2", 0},
		{"Window Display", "offline", "192.168.1.102", "1.3.9", 0},
		{"Lobby Screen", "online", "192.168.2.10", "1.4.2", 1},
		{"Cardio Zone Screen", "online", "192.168.2.11", "1.4.1", 1},
		{"Entrance Board", "online", "192.168.3.50", "1.4.2", 2},
		{"Main Atrium Screen", "online", "192.168.4.20", "1.4.2", 3},
		{"Food Court Display", "offline", "192.168.4.21", "1.3.9", 3},
	}

	// days: 0=Sun 1=Mon … 6=Sat
	weekdays := []int{1, 2, 3, 4, 5}
	weekend := []int{0, 6}
	allWeek := []int{0, 1, 2, 3, 4, 5, 6}

	slots := []slotRow{
		{"Morning Rush", weekdays, "07:00", "10:00", 15, 4500, "booked", 0},
		{"Lunch Hour", allWeek, "11:30", "13:30", 15, 6000, "booked", 0},
		{"Afternoon Lull", weekdays, "14:00", "17:00", 15, 3000, "available", 0},
		{"Evening", allWeek, "17:00", "20:00", 15, 5000, "available", 1},
		{"Morning Workout", weekdays, "06:00", "09:00", 15, 5500, "booked", 2},
		{"Midday", allWeek, "11:00", "14:00", 15, 4000, "available", 2},
		{"Evening Rush", weekdays, "17:00", "20:00", 15, 6500, "booked", 3},
		{"Weekend Afternoon", weekend, "12:00", "18:00", 15, 7000, "booked", 4},
		{"Weekday Prime", weekdays, "08:00", "20:00", 30, 12000, "booked", 5},
		{"Weekend All-Day", weekend, "10:00", "21:00", 30, 15000, "available", 6},
	}

	bookings := []bookingRow{
		{0, 0, "2026-04-01", "2026-06-30", 4500, "active"},
		{1, 2, "2026-05-01", "2026-05-31", 6000, "active"},
		{4, 1, "2026-03-01", "2026-03-31", 5500, "completed"},
		{6, 3, "2026-05-01", "2026-07-31", 6500, "active"},
		{7, 2, "2026-06-01", "2026-06-30", 7000, "pending"},
		{8, 0, "2026-05-13", "2026-08-31", 12000, "active"},
		{1, 3, "2026-07-01", "2026-09-30", 6000, "pending"},
	}

	// Insert venues
	venueIDs := make([]string, len(venues))
	for i, v := range venues {
		var id string
		err := pool.QueryRow(ctx, `SELECT id FROM venues WHERE name = $1 LIMIT 1`, v.name).Scan(&id)
		if err != nil {
			err = pool.QueryRow(ctx,
				`INSERT INTO venues (name, category, address, city, state)
				 VALUES ($1, $2, $3, $4, $5) RETURNING id`,
				v.name, v.category, v.address, v.city, v.state,
			).Scan(&id)
			if err != nil {
				return fmt.Errorf("insert venue %q: %w", v.name, err)
			}
			fmt.Printf("✓ venue created       %s\n", v.name)
		} else {
			fmt.Printf("✓ venue exists        %s\n", v.name)
		}
		venueIDs[i] = id
	}

	// Insert advertisers
	advertiserIDs := make([]string, len(advertisers))
	for i, a := range advertisers {
		var id string
		err := pool.QueryRow(ctx, `SELECT id FROM advertisers WHERE email = $1 LIMIT 1`, a.email).Scan(&id)
		if err != nil {
			err = pool.QueryRow(ctx,
				`INSERT INTO advertisers (name, email) VALUES ($1, $2) RETURNING id`,
				a.name, a.email,
			).Scan(&id)
			if err != nil {
				return fmt.Errorf("insert advertiser %q: %w", a.name, err)
			}
			fmt.Printf("✓ advertiser created  %s\n", a.name)
		} else {
			fmt.Printf("✓ advertiser exists   %s\n", a.name)
		}
		advertiserIDs[i] = id
	}

	// Insert devices
	deviceIDs := make([]string, len(devices))
	for i, d := range devices {
		var id string
		err := pool.QueryRow(ctx, `SELECT id FROM devices WHERE name = $1 AND venue_id = $2 LIMIT 1`,
			d.name, venueIDs[d.venueIdx]).Scan(&id)
		if err != nil {
			err = pool.QueryRow(ctx,
				`INSERT INTO devices (venue_id, name, status, ip_address, firmware_version)
				 VALUES ($1, $2, $3, $4, $5) RETURNING id`,
				venueIDs[d.venueIdx], d.name, d.status, d.ip, d.firmware,
			).Scan(&id)
			if err != nil {
				return fmt.Errorf("insert device %q: %w", d.name, err)
			}
			fmt.Printf("✓ device created      %s\n", d.name)
		} else {
			fmt.Printf("✓ device exists       %s\n", d.name)
		}
		deviceIDs[i] = id
	}

	// Insert slots
	slotIDs := make([]string, len(slots))
	for i, s := range slots {
		var id string
		err := pool.QueryRow(ctx, `SELECT id FROM ad_slots WHERE label = $1 AND device_id = $2 LIMIT 1`,
			s.label, deviceIDs[s.deviceIdx]).Scan(&id)
		if err != nil {
			err = pool.QueryRow(ctx,
				`INSERT INTO ad_slots (device_id, label, days_of_week, start_time, end_time, duration_sec, price_cents, status)
				 VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING id`,
				deviceIDs[s.deviceIdx], s.label, s.daysOfWeek, s.startTime, s.endTime,
				s.durationSec, s.priceCents, s.status,
			).Scan(&id)
			if err != nil {
				return fmt.Errorf("insert slot %q: %w", s.label, err)
			}
			fmt.Printf("✓ slot created        %s — %s\n", s.label, s.startTime)
		} else {
			fmt.Printf("✓ slot exists         %s — %s\n", s.label, s.startTime)
		}
		slotIDs[i] = id
	}

	// Insert bookings
	for _, b := range bookings {
		var id string
		err := pool.QueryRow(ctx,
			`SELECT id FROM bookings WHERE slot_id = $1 AND advertiser_id = $2 AND starts_on = $3 LIMIT 1`,
			slotIDs[b.slotIdx], advertiserIDs[b.advertiserIdx], b.startsOn,
		).Scan(&id)
		if err != nil {
			err = pool.QueryRow(ctx,
				`INSERT INTO bookings (slot_id, advertiser_id, starts_on, ends_on, price_cents, status)
				 VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`,
				slotIDs[b.slotIdx], advertiserIDs[b.advertiserIdx],
				b.startsOn, b.endsOn, b.priceCents, b.status,
			).Scan(&id)
			if err != nil {
				return fmt.Errorf("insert booking: %w", err)
			}
			fmt.Printf("✓ booking created     slot[%d] × advertiser[%d] → %s\n", b.slotIdx, b.advertiserIdx, b.status)
		} else {
			fmt.Printf("✓ booking exists      slot[%d] × advertiser[%d] → %s\n", b.slotIdx, b.advertiserIdx, b.status)
		}
	}

	return nil
}
