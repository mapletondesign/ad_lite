package advertisers

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

func (s *Service) Create(ctx context.Context, req CreateRequest, actorID string) (Advertiser, error) {
	if req.Name == "" || req.Email == "" {
		return Advertiser{}, fmt.Errorf("name and email are required")
	}
	var createdBy *string
	if actorID != "" {
		createdBy = &actorID
	}
	var a Advertiser
	err := s.db.QueryRow(ctx, `
		INSERT INTO advertisers (name, email, created_by)
		VALUES ($1, $2, $3)
		RETURNING id, name, email, created_at`,
		req.Name, req.Email, createdBy,
	).Scan(&a.ID, &a.Name, &a.Email, &a.CreatedAt)
	if err != nil {
		return Advertiser{}, fmt.Errorf("insert advertiser: %w", err)
	}
	return a, nil
}

func (s *Service) Get(ctx context.Context, id string) (Advertiser, error) {
	var a Advertiser
	err := s.db.QueryRow(ctx, `
		SELECT id, name, email, created_at FROM advertisers WHERE id = $1`, id,
	).Scan(&a.ID, &a.Name, &a.Email, &a.CreatedAt)
	if err != nil {
		return Advertiser{}, fmt.Errorf("advertiser not found")
	}
	return a, nil
}

func (s *Service) List(ctx context.Context) ([]Advertiser, error) {
	rows, err := s.db.Query(ctx, `
		SELECT id, name, email, created_at FROM advertisers ORDER BY created_at DESC`)
	if err != nil {
		return nil, fmt.Errorf("query advertisers: %w", err)
	}
	defer rows.Close()

	var advertisers []Advertiser
	for rows.Next() {
		var a Advertiser
		if err := rows.Scan(&a.ID, &a.Name, &a.Email, &a.CreatedAt); err != nil {
			return nil, err
		}
		advertisers = append(advertisers, a)
	}
	return advertisers, rows.Err()
}
