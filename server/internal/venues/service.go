package venues

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

func (s *Service) Create(ctx context.Context, req CreateRequest) (Venue, error) {
	if req.Name == "" {
		return Venue{}, fmt.Errorf("name is required")
	}
	var v Venue
	err := s.db.QueryRow(ctx, `
		INSERT INTO venues (name, category, address, city, state)
		VALUES ($1, NULLIF($2,''), NULLIF($3,''), NULLIF($4,''), NULLIF($5,''))
		RETURNING id, name, category, address, city, state, country, created_at`,
		req.Name, req.Category, req.Address, req.City, req.State,
	).Scan(&v.ID, &v.Name, &v.Category, &v.Address, &v.City, &v.State, &v.Country, &v.CreatedAt)
	if err != nil {
		return Venue{}, fmt.Errorf("insert venue: %w", err)
	}
	return v, nil
}

func (s *Service) List(ctx context.Context) ([]Venue, error) {
	rows, err := s.db.Query(ctx, `
		SELECT id, name, category, address, city, state, country, created_at
		FROM venues ORDER BY created_at DESC`)
	if err != nil {
		return nil, fmt.Errorf("query venues: %w", err)
	}
	defer rows.Close()

	var venues []Venue
	for rows.Next() {
		var v Venue
		if err := rows.Scan(&v.ID, &v.Name, &v.Category, &v.Address, &v.City, &v.State, &v.Country, &v.CreatedAt); err != nil {
			return nil, err
		}
		venues = append(venues, v)
	}
	return venues, rows.Err()
}
