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
	demo := flag.Bool("demo", false, "also seed a demo venue, advertiser, and device")
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
	err = pool.QueryRow(ctx,
		`SELECT id FROM users WHERE email = $1`, email,
	).Scan(&id)
	if err == nil {
		return id, false, nil
	}

	hash, err := auth.HashPassword(password)
	if err != nil {
		return "", false, fmt.Errorf("hash password: %w", err)
	}

	err = pool.QueryRow(ctx,
		`INSERT INTO users (email, password_hash, role)
		 VALUES ($1, $2, 'admin')
		 RETURNING id`,
		email, hash,
	).Scan(&id)
	if err != nil {
		return "", false, fmt.Errorf("insert user: %w", err)
	}
	return id, true, nil
}

func seedDemo(ctx context.Context, pool *pgxpool.Pool) error {
	var venueID string
	err := pool.QueryRow(ctx, `SELECT id FROM venues WHERE name = 'Demo Venue' LIMIT 1`).Scan(&venueID)
	if err != nil {
		err = pool.QueryRow(ctx,
			`INSERT INTO venues (name, category, address, city, state)
			 VALUES ('Demo Venue', 'retail', '123 Main St', 'Springfield', 'IL')
			 RETURNING id`,
		).Scan(&venueID)
		if err != nil {
			return fmt.Errorf("insert venue: %w", err)
		}
		fmt.Printf("✓ venue created  name=Demo Venue  id=%s\n", venueID)
	} else {
		fmt.Printf("✓ venue exists   name=Demo Venue  id=%s\n", venueID)
	}

	var advertiserID string
	err = pool.QueryRow(ctx, `SELECT id FROM advertisers WHERE email = 'demo@advertiser.com' LIMIT 1`).Scan(&advertiserID)
	if err != nil {
		err = pool.QueryRow(ctx,
			`INSERT INTO advertisers (name, email)
			 VALUES ('Demo Advertiser', 'demo@advertiser.com')
			 RETURNING id`,
		).Scan(&advertiserID)
		if err != nil {
			return fmt.Errorf("insert advertiser: %w", err)
		}
		fmt.Printf("✓ advertiser created  name=Demo Advertiser  id=%s\n", advertiserID)
	} else {
		fmt.Printf("✓ advertiser exists   name=Demo Advertiser  id=%s\n", advertiserID)
	}

	var deviceID string
	err = pool.QueryRow(ctx, `SELECT id FROM devices WHERE name = 'Demo Screen' LIMIT 1`).Scan(&deviceID)
	if err != nil {
		err = pool.QueryRow(ctx,
			`INSERT INTO devices (venue_id, name, status)
			 VALUES ($1, 'Demo Screen', 'offline')
			 RETURNING id`,
			venueID,
		).Scan(&deviceID)
		if err != nil {
			return fmt.Errorf("insert device: %w", err)
		}
		fmt.Printf("✓ device created  name=Demo Screen  id=%s\n", deviceID)
	} else {
		fmt.Printf("✓ device exists   name=Demo Screen  id=%s\n", deviceID)
	}

	return nil
}
