package auth

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

const refreshTokenDuration = 7 * 24 * time.Hour

type Service struct {
	db         *pgxpool.Pool
	privateKey *rsa.PrivateKey
	publicKey  *rsa.PublicKey
}

func NewService(db *pgxpool.Pool, privateKey *rsa.PrivateKey, publicKey *rsa.PublicKey) *Service {
	return &Service{db: db, privateKey: privateKey, publicKey: publicKey}
}

func (s *Service) Login(ctx context.Context, req LoginRequest) (TokenPair, error) {
	var u User
	err := s.db.QueryRow(ctx,
		`SELECT id, email, password_hash, role, advertiser_id::text, venue_id::text
		 FROM users WHERE email = $1`,
		req.Email,
	).Scan(&u.ID, &u.Email, &u.PasswordHash, &u.Role, &u.AdvertiserID, &u.VenueID)
	if err != nil {
		return TokenPair{}, fmt.Errorf("invalid credentials")
	}
	if err := CheckPassword(u.PasswordHash, req.Password); err != nil {
		return TokenPair{}, fmt.Errorf("invalid credentials")
	}
	return s.issueTokenPair(ctx, &u)
}

func (s *Service) Refresh(ctx context.Context, req RefreshRequest) (TokenPair, error) {
	hash := hashToken(req.RefreshToken)

	var tokenID, userID string
	var expiresAt time.Time
	err := s.db.QueryRow(ctx,
		`SELECT id, user_id, expires_at FROM refresh_tokens WHERE token_hash = $1`,
		hash,
	).Scan(&tokenID, &userID, &expiresAt)
	if err != nil || time.Now().After(expiresAt) {
		return TokenPair{}, fmt.Errorf("invalid or expired refresh token")
	}

	var u User
	err = s.db.QueryRow(ctx,
		`SELECT id, email, role, advertiser_id::text, venue_id::text FROM users WHERE id = $1`,
		userID,
	).Scan(&u.ID, &u.Email, &u.Role, &u.AdvertiserID, &u.VenueID)
	if err != nil {
		return TokenPair{}, fmt.Errorf("user not found")
	}

	if _, err := s.db.Exec(ctx, `DELETE FROM refresh_tokens WHERE id = $1`, tokenID); err != nil {
		return TokenPair{}, fmt.Errorf("rotate refresh token: %w", err)
	}

	return s.issueTokenPair(ctx, &u)
}

func (s *Service) Register(ctx context.Context, req RegisterRequest) (TokenPair, error) {
	if req.Email == "" || req.Password == "" || req.Role == "" || req.Name == "" {
		return TokenPair{}, fmt.Errorf("email, password, role, and name are required")
	}
	if req.Role != "advertiser" && req.Role != "venue" {
		return TokenPair{}, fmt.Errorf("role must be 'advertiser' or 'venue'")
	}

	hash, err := HashPassword(req.Password)
	if err != nil {
		return TokenPair{}, fmt.Errorf("hash password: %w", err)
	}

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return TokenPair{}, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	var u User
	u.Role = req.Role

	switch req.Role {
	case "advertiser":
		var id string
		if err := tx.QueryRow(ctx,
			`INSERT INTO advertisers (name, email) VALUES ($1, $2) RETURNING id`,
			req.Name, req.Email,
		).Scan(&id); err != nil {
			return TokenPair{}, fmt.Errorf("create advertiser: %w", err)
		}
		u.AdvertiserID = &id
		if err := tx.QueryRow(ctx,
			`INSERT INTO users (email, password_hash, role, advertiser_id) VALUES ($1, $2, $3, $4) RETURNING id`,
			req.Email, hash, req.Role, id,
		).Scan(&u.ID); err != nil {
			return TokenPair{}, fmt.Errorf("create user: %w", err)
		}

	case "venue":
		var id string
		if err := tx.QueryRow(ctx,
			`INSERT INTO venues (name) VALUES ($1) RETURNING id`,
			req.Name,
		).Scan(&id); err != nil {
			return TokenPair{}, fmt.Errorf("create venue: %w", err)
		}
		u.VenueID = &id
		if err := tx.QueryRow(ctx,
			`INSERT INTO users (email, password_hash, role, venue_id) VALUES ($1, $2, $3, $4) RETURNING id`,
			req.Email, hash, req.Role, id,
		).Scan(&u.ID); err != nil {
			return TokenPair{}, fmt.Errorf("create user: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return TokenPair{}, fmt.Errorf("commit: %w", err)
	}

	return s.issueTokenPair(ctx, &u)
}

func (s *Service) issueTokenPair(ctx context.Context, u *User) (TokenPair, error) {
	advertiserID, venueID := "", ""
	if u.AdvertiserID != nil {
		advertiserID = *u.AdvertiserID
	}
	if u.VenueID != nil {
		venueID = *u.VenueID
	}

	accessToken, err := IssueAccessToken(s.privateKey, u.ID, u.Role, advertiserID, venueID)
	if err != nil {
		return TokenPair{}, err
	}

	rawRefresh, err := generateRawToken()
	if err != nil {
		return TokenPair{}, fmt.Errorf("generate refresh token: %w", err)
	}

	if _, err := s.db.Exec(ctx,
		`INSERT INTO refresh_tokens (user_id, token_hash, expires_at) VALUES ($1, $2, $3)`,
		u.ID, hashToken(rawRefresh), time.Now().Add(refreshTokenDuration),
	); err != nil {
		return TokenPair{}, fmt.Errorf("store refresh token: %w", err)
	}

	return TokenPair{
		AccessToken:  accessToken,
		RefreshToken: rawRefresh,
		ExpiresIn:    int(AccessTokenDuration.Seconds()),
	}, nil
}

func generateRawToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func hashToken(raw string) string {
	h := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(h[:])
}
