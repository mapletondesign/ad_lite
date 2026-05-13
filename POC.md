# AdSling — Proof of Concept

**Goal:** A single display device pulls a playlist from a local server and rotates ads on screen. No billing, no marketplace, no mobile app — just prove the core loop works end to end.

**Definition of done:** A Raspberry Pi (or any laptop standing in for one) boots, contacts the server, downloads a playlist of two ads, and cycles them on screen at the configured interval. The server records a play confirmation event each time an ad is displayed.

---

## What You Need

- Raspberry Pi 5 (or any laptop/desktop for local dev)
- HDMI display or second monitor
- Go installed (1.22+)
- Node.js 20+ (for the Vue web client running in kiosk mode)
- Two ad creatives: one 1920×1080 JPEG and one short MP4 (under 30s)

---

## Step 1 — Scaffold the server

Create `ad_sling/server/` as a Go module.

```
ad_sling/server/
├── go.mod
├── main.go
├── handlers/
│   ├── playlist.go    # GET /api/playlist
│   └── events.go      # POST /api/events
└── data/
    └── ads.go         # in-memory ad store
```

The server needs exactly two endpoints for the POC:

**`GET /api/playlist?device_id=<id>`**
Returns a JSON array of ads to display, in order:
```json
[
  { "id": "ad_001", "type": "image", "url": "/assets/ad1.jpg", "duration_s": 10 },
  { "id": "ad_002", "type": "video", "url": "/assets/ad2.mp4", "duration_s": 15 }
]
```

**`POST /api/events`**
Accepts a play confirmation and logs it to stdout:
```json
{ "device_id": "pi-01", "ad_id": "ad_001", "played_at": 1714400000 }
```

**`GET /assets/*`**
Static file handler serving ad creatives from `server/assets/`.

No database for the POC — hardcode two ads in `data/ads.go`. Keep the server on port `8080`.

---

## Step 2 — Add your two ad files

Copy your JPEG and MP4 into `ad_sling/server/assets/`:

```
server/assets/
├── ad1.jpg
└── ad2.mp4
```

---

## Step 3 — Build and run the server

```bash
cd ad_sling/server
go mod tidy
go run .
```

Verify manually:
```bash
curl http://localhost:8080/api/playlist?device_id=pi-01
```
You should get the two-ad JSON array back.

---

## Step 4 — Scaffold the kiosk client

Create `ad_sling/client/` as a plain HTML5 + vanilla JS app. No framework needed for the POC.

```
ad_sling/client/
├── index.html
├── css/
│   └── player.css
└── js/
    └── player.js
```

`index.html` — full-screen black background, a single `<div id="player">` containing either an `<img>` or a `<video>` element.

`js/player.js` — on load:
1. Fetch `GET /api/playlist?device_id=pi-01`
2. Build a local queue from the response
3. Show the first ad (set `<img src>` or `<video src>`, play video if type is `video`)
4. When the ad's `duration_s` elapses (or the video `ended` event fires), POST to `/api/events`, then advance to the next ad in the queue
5. Loop back to step 1 when the queue is exhausted (re-fetch the playlist so the server can update it live)

No bundler, no build step — just files the browser reads directly.

---

## Step 5 — Wire server to serve the client

Add a static handler in `main.go` to serve `../client/` at `/`. Now the server handles both the API and the player UI from one binary:

- `http://localhost:8080/` → kiosk player
- `http://localhost:8080/api/playlist` → playlist endpoint
- `http://localhost:8080/assets/` → ad creatives

---

## Step 6 — Open in kiosk mode

On the Pi (or any dev machine), open Chromium in kiosk mode pointed at the local server:

```bash
chromium-browser --kiosk --noerrdialogs --disable-infobars \
  --app=http://localhost:8080
```

On macOS for local dev:
```bash
open -a "Google Chrome" --args --kiosk http://localhost:8080
```

You should see the first ad fill the screen, transition to the second after its duration, then loop.

---

## Step 7 — Confirm event logging

While the player is running, watch the server terminal. Each time an ad completes you should see a log line like:

```
PLAY  device=pi-01  ad=ad_001  at=2026-04-29T14:32:00Z
```

This confirms the full loop: schedule delivered → ad displayed → impression recorded.

---

## Step 8 — Test a live playlist update

While the player is running, edit `data/ads.go` on the server to change the duration of one ad or swap in a different asset. Restart the server. The client will re-fetch the playlist on its next loop cycle and pick up the change without touching the player.

---

## POC Complete

You have proven:
- A Go server can deliver a playlist to a device and receive play confirmations
- A browser-based kiosk player can rotate mixed media (image + video) from a remote schedule
- The playlist is hot-updatable without touching the player

**What is explicitly out of scope for the POC (covered in PLAN.md):**
- Device registration or authentication
- PostgreSQL persistence
- S3/CDN asset storage
- Billing or venue revenue share
- Scheduling by time-of-day or targeting
- Web portal UI
- Mobile app
- OTA updates or watchdog
