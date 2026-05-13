# AdSling — Pre-Launch Checklist

Check every item before putting real devices or advertisers on the platform. Grouped by area.

---

## Security — Critical (blocking)

- [x] Replace UUID device tokens with signed JWTs (HS256, upgrade to RS256 when multi-service)
- [x] Add JWT auth middleware — device routes require valid device JWT
- [x] Add API key auth — management routes require `X-API-Key` header (stopgap until Stage 4 user auth)
- [ ] Implement full role enforcement: `advertiser`, `venue`, `admin` (Stage 4)
- [x] Stop returning raw Go error strings to clients — generic 500 message + server-side log
- [x] Cap request body size with `http.MaxBytesReader` (1MB) on all routes
- [ ] Validate `creative_url` against CDN domain allowlist before saving
- [ ] Validate UUID path params before DB queries (reject malformed IDs early)
- [ ] Verify device exists before accepting impression records

## Security — High (before launch)

- [x] Add security headers middleware (`X-Content-Type-Options`, `X-Frame-Options`, `HSTS`, `Referrer-Policy`)
- [ ] Configure CORS — allowlist known origins, never wildcard
- [ ] Add rate limiting on `/heartbeat` (2/min per device) and `/impression` (10/min per device)
- [ ] Add rate limiting on registration and auth endpoints (per IP)
- [ ] Add audit fields (`created_by`, `updated_by`) to bookings, venues, advertisers
- [ ] Add admin action audit log

## Infrastructure — Critical (blocking)

- [ ] Move ad creative serving to S3/Cloudflare R2 + CDN — remove local filesystem serving
- [ ] Implement presigned S3 upload URLs — clients never receive S3 credentials directly
- [ ] Cache playlist results in Redis per device (TTL ~60s)
- [ ] Deploy behind HTTPS — enforce `Strict-Transport-Security`
- [ ] Move secrets to a secrets manager or CI environment variables (not `.env` files in production)

## Infrastructure — High

- [ ] Add Redis-backed rate limiting (Redis is connected but unused)
- [ ] Set up database backups (automated, daily minimum)
- [ ] Set up uptime monitoring and alerting (device offline detection)
- [ ] Configure graceful error alerting (Sentry or equivalent)
- [ ] Run `go mod verify` and `npm audit` in CI pipeline

## API Completeness

- [ ] `POST /auth/device/login` — device exchanges token for JWT
- [ ] `POST /auth/refresh` — refresh expired access token
- [ ] `POST /auth/register` — advertiser/venue self-registration
- [ ] `PATCH /bookings/{id}` — cancel or update a booking
- [ ] `DELETE /slots/{id}` — remove a slot
- [ ] `GET /analytics/impressions` — currently a stub, implement before launch
- [ ] `GET /venues/{id}` and `GET /advertisers/{id}` — single-record fetch endpoints

## Data Integrity

- [ ] Add `CHECK` constraint: `ends_on >= starts_on` on bookings table
- [ ] Add `CHECK` constraint: `end_time > start_time` on ad_slots table
- [ ] Add `CHECK` constraint: `price_cents > 0` on ad_slots and bookings
- [ ] Add `CHECK` constraint: `duration_sec > 0` on ad_slots
- [ ] Add `CHECK` constraint: `days_of_week` values are 0–6 only
- [ ] Migration for `updated_at` trigger on all tables that have that column

## Testing

- [ ] Integration tests for booking conflict/availability logic
- [ ] Integration tests for scheduler playlist query
- [ ] Load test heartbeat endpoint at 10× expected device count
- [ ] End-to-end test: full flow from device registration to ad playback to impression
- [ ] Test device behaviour when server returns empty playlist
- [ ] Test device behaviour when server is unreachable

## Operational

- [ ] Define and document device deprovisioning flow (token rotation, slot reassignment)
- [ ] Define content moderation process for ad creatives
- [ ] Confirm legal review of digital signage regulations for target markets
- [ ] Confirm venue revenue share payout mechanism (Stripe Connect or manual)
- [ ] Hardware provisioning runbook — how to set up a new Pi 5 for a venue
