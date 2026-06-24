package audit

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Service struct {
	db *pgxpool.Pool
}

func NewService(db *pgxpool.Pool) *Service {
	return &Service{db: db}
}

// Log records one admin action. It is fire-and-forget: failures are logged to
// stderr but never returned to the caller.
func (s *Service) Log(actorID, action, targetTable, targetID string, details map[string]any) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var detailsJSON []byte
	if details != nil {
		detailsJSON, _ = json.Marshal(details)
	}

	var tid *string
	if targetID != "" {
		tid = &targetID
	}

	if _, err := s.db.Exec(ctx, `
		INSERT INTO admin_audit_log (actor_id, action, target_table, target_id, details)
		VALUES ($1, $2, $3, $4, $5)`,
		actorID, action, targetTable, tid, detailsJSON,
	); err != nil {
		log.Printf("audit log: %v", err)
	}
}
