-- Migration 005 — audit fields and admin action log

ALTER TABLE venues
  ADD COLUMN created_by uuid REFERENCES users(id) ON DELETE SET NULL,
  ADD COLUMN updated_by uuid REFERENCES users(id) ON DELETE SET NULL;

ALTER TABLE advertisers
  ADD COLUMN created_by uuid REFERENCES users(id) ON DELETE SET NULL,
  ADD COLUMN updated_by uuid REFERENCES users(id) ON DELETE SET NULL;

ALTER TABLE bookings
  ADD COLUMN created_by uuid REFERENCES users(id) ON DELETE SET NULL,
  ADD COLUMN updated_by uuid REFERENCES users(id) ON DELETE SET NULL;

CREATE TABLE admin_audit_log (
  id           bigserial   PRIMARY KEY,
  actor_id     uuid        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  action       text        NOT NULL,
  target_table text        NOT NULL,
  target_id    text,
  details      jsonb,
  created_at   timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX idx_admin_audit_log_actor_id   ON admin_audit_log(actor_id);
CREATE INDEX idx_admin_audit_log_created_at ON admin_audit_log(created_at DESC);
