# AdLite

A digital out-of-home (DOOH) advertising platform. Venues such as restaurants, offices and gyms host a screen driven by a Raspberry Pi 5 running a kiosk player. Advertisers book time slots on those screens, and venues earn a share of the ad spend.

## How it works

1. A venue registers a device, which receives a signed device token.
2. The device runs the kiosk player, which fetches its playlist, rotates image and video ads, and sends a heartbeat every 60 seconds.
3. Each time an ad is shown, the player reports an impression to the API.
4. Advertisers upload creatives, book slots and track impressions in the web portal. Admins manage venues, devices, slots and bookings.

## Repository layout

| Path | What it is |
|---|---|
| `server/` | Go API (chi v5, pgx/v5, Redis). Domain packages live under `internal/`: auth, devices, slots, bookings, scheduler, analytics, assets, audit. |
| `server/migrations/` | Postgres schema, applied automatically when the database container first boots. |
| `client/` | Kiosk player in vanilla HTML and JavaScript. The API serves it at `/`. |
| `portal/` | Nuxt 3 web portal (Vue 3, Vuetify, Pinia) with advertiser, venue and admin areas. |
| `docs/` | HTML reference pages: local dev guide, system flow, personas and workflows. |
| `PLAN.md` | Product plan, stages and unit economics. |
| `SECURITY.md` | Security patterns the codebase follows. |
| `LAUNCH_CHECKLIST.md` | Work outstanding before real devices go live. |

## Requirements

- Go 1.26+
- Node.js 20+ (Node 24 to run the player tests)
- Docker, for Postgres 16 and Redis 7

## Running locally

### 1. Configure the server

```bash
cd server
cp .env.example .env
```

Generate the RSA keypair used to sign JWTs and append it to `.env`. Do this once; regenerating invalidates existing tokens.

```bash
openssl genrsa -out /tmp/adlite.pem 2048
openssl rsa -in /tmp/adlite.pem -pubout -out /tmp/adlite.pub
echo "JWT_PRIVATE_KEY_B64=$(base64 -i /tmp/adlite.pem)" >> .env
echo "JWT_PUBLIC_KEY_B64=$(base64 -i /tmp/adlite.pub)" >> .env
rm /tmp/adlite.pem /tmp/adlite.pub
```

Stripe and S3/R2 values are optional for local work. If the S3 variables are unset, the upload endpoint is disabled at startup.

### 2. Start Postgres and Redis

```bash
make up
```

### 3. Seed an admin user

```bash
go run ./cmd/seed --password <admin-password> -demo
```

`--password` is required. `--email` defaults to `admin@adlite.com`. `-demo` also creates sample venues, advertisers, devices, slots and bookings. The seed is safe to re-run.

### 4. Start the API

```bash
make dev
```

The API listens on `http://localhost:8080`, with a health check at `/health` and the kiosk player at `/`.

### 5. Start the portal

```bash
cd ../portal
npm install
npm run dev
```

The portal runs on `http://localhost:3000`.

## API overview

All routes are under `/api/v1`. Access is enforced by role in middleware.

| Area | Routes |
|---|---|
| Auth | `POST /auth/register`, `/auth/login`, `/auth/refresh`, `/auth/device/login` |
| Admin | `/venues`, `/advertisers`, `/slots`, `/bookings`, `/devices`, `/analytics/impressions` |
| Advertiser | `/advertiser/slots`, `/advertiser/bookings`, `/advertiser/analytics/impressions` |
| Venue | `/venue/devices` |
| Device | `POST /devices/register`, `/devices/{id}/heartbeat`, `/devices/{id}/impression` |
| Assets | `GET /assets/presign`, `POST /assets/upload` |

See `server/api/router.go` for the full list.

## Testing

```bash
# Server: needs Postgres and Redis running (make up)
cd server
make test
make lint

# Kiosk player
node --test client/tests/player.test.mjs

# Portal type check
cd portal
npm run typecheck
```

Server tests run sequentially because every package shares one `ad_lite_test` database.

## Status

The backend, kiosk player and web portal are built. A mobile companion app, a self-serve marketplace with Stripe checkout, and the pre-launch hardening items are still to come. `PLAN.md` and `LAUNCH_CHECKLIST.md` have the detail.
