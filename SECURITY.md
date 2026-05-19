# AdLite — Security Patterns

Reference for all security decisions. Follow these patterns consistently across every stage.

---

## Authentication

- **Device auth:** JWT signed with RS256. Issued at registration, stored on device. Short expiry (24h) with refresh. Verified on every API call via middleware.
- **User auth (web/mobile):** JWT — 15min access token, 7-day refresh token. RS256. Never HS256 in a multi-service setup.
- **Roles:** `device`, `advertiser`, `venue`, `admin`. Enforced in middleware, not in individual handlers.
- **Never:** random UUIDs as auth tokens, plaintext tokens in logs, tokens in URLs.

## Input Validation

- Validate at the HTTP handler boundary — reject early, never sanitize-and-continue.
- All UUID path params must be validated as UUIDs before hitting the DB.
- `creative_url` must match an allowlisted domain pattern (e.g. your own S3/CDN bucket). Never accept arbitrary URLs from clients.
- Request bodies must be capped at a max size (use `http.MaxBytesReader`).
- Date strings (`starts_on`, `ends_on`) must parse cleanly — reject anything that doesn't.

## Error Responses

- **Never** return raw Go error strings to clients. They expose DB schema and internal structure.
- Public error format: `{ "error": "<human message>", "request_id": "<id>" }` only.
- Log the full internal error server-side with the request ID for correlation.
- 500s get a generic "internal server error" message — never the underlying cause.

## SQL

- Parameterized queries only. Already enforced via pgx. Never string-concatenate SQL.
- Use transactions for any multi-step write (already done in bookings). 

## Security Headers

Add to all responses via middleware:

```
X-Content-Type-Options: nosniff
X-Frame-Options: DENY
Referrer-Policy: strict-origin-when-cross-origin
Strict-Transport-Security: max-age=63072000; includeSubDomains
Content-Security-Policy: default-src 'self'
```

## CORS

- API routes: allow only known origins (web portal domain, admin domain). Never `*`.
- Device routes (`/heartbeat`, `/impression`): no CORS needed — device-to-server only.
- Configure per-environment via env vars, not hardcoded.

## Rate Limiting

- Per-device limits on `/heartbeat` (max 2/min) and `/impression` (max 10/min).
- Per-IP limits on registration and auth endpoints.
- Backed by Redis. Return `429 Too Many Requests` with `Retry-After` header.

## Secrets

- Never commit `.env`, credentials, or private keys.
- JWT private key loaded from env var or secrets manager — never from a file in the repo.
- Rotate tokens if a device is deprovisioned.
- Run `go mod verify` in CI.

## Asset Security

- Ad creatives stored in S3/R2, served via CDN.
- `creative_url` stored in DB must match allowlisted CDN domain before saving.
- Presigned S3 upload URLs issued server-side — clients never get direct S3 credentials.

## Audit Log

- Record `created_by`, `updated_by` on bookings, venues, advertisers.
- Log all admin actions (approve venue, override schedule, cancel booking).
