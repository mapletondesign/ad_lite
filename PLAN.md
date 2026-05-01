# AdPack — Digital Out-of-Home (DOOH) Advertising Platform
## Project Plan

---

## Overview

AdPack is a digital out-of-home advertising network that places billboard-quality ads on venue screens (restaurant TVs, office monitors, waiting rooms, retail displays). Venues get free or subsidized hardware and a revenue share; advertisers get targeted, location-aware placements at a fraction of traditional billboard cost.

---

## Locked Decisions (2026-05-01)

| Decision | Choice |
|---|---|
| Backend language | Go |
| Pricing model v1 | Time-slot flat rate |
| Pricing model v2 | CPM (added once impression tracking is solid) |
| Hardware | Raspberry Pi 5 (4GB) for dev and pilot |

---

## Stage Breakdown

---

### Stage 1 — Hardware Selection & Specification

**Goal:** Define the hardware stack for venue displays.

**Recommended Device:** Raspberry Pi 5 (4GB) or equivalent ARM SBC, or a commercial Android media player (BrightSign, Zeetaminds, or a low-cost Android stick like the UGOOS AM6B+).

**Evaluation Criteria:**
- Must run a locked-down kiosk OS or managed Android
- HDMI output at 1080p/4K
- Persistent network connection (Ethernet preferred, WiFi fallback)
- Remote management capability (reboot, update, status)
- Boot-to-app with no user interaction required
- Industrial temperature range preferred for kitchen/retail environments
- Cost target: **< $80/unit at scale**

**Leading Options:**

| Device | Cost | OS | Notes |
|---|---|---|---|
| Raspberry Pi 5 (4GB) | ~$60 | Raspberry Pi OS / custom Linux | Most flexible, active community |
| BrightSign HD5 | ~$300 | BrightOS | Enterprise-grade, expensive |
| UGOOS AM6B+ | ~$80 | Android 9 | Good price/performance |
| Fire TV Stick 4K | ~$50 | Android fork | Limited remote mgmt |

**Recommended:** Start with Raspberry Pi 5 for development; evaluate Android stick (UGOOS or similar) for production cost reduction.

**Hardware Bundle per Venue:**
- 1x media player device
- 1x HDMI cable + power adapter
- Optional: wall mount bracket
- Pre-flashed SD card / pre-provisioned device

**Deliverables:**
- [ ] Finalize device selection
- [ ] Procure 10 pilot units
- [ ] Define hardware provisioning workflow

---

### Stage 2 — Main Server (Backend Infrastructure)

**Goal:** Build the central platform that manages devices, serves ad content, and handles billing.

**Tech Stack (Recommended):**
- **Runtime:** Node.js (TypeScript) or Go
- **API:** REST + WebSocket for device heartbeat/commands
- **Database:** PostgreSQL (primary), Redis (caching, pub/sub)
- **File Storage:** S3-compatible (AWS S3 or Cloudflare R2 for cost)
- **CDN:** Cloudflare or AWS CloudFront for ad asset delivery
- **Hosting:** AWS / GCP / Hetzner (Hetzner for cost efficiency in early stages)
- **Auth:** JWT for client devices, OAuth2 for web/mobile users

**Core Services:**

1. **Device Management Service**
   - Register/provision devices
   - Track heartbeat and online status
   - Push configuration and schedule updates
   - Remote commands (reboot, force refresh)

2. **Ad Scheduling Engine**
   - Schedule ads by time-of-day, day-of-week, venue, region
   - Playlist builder: rotation weights, frequency caps
   - Real-time slot assignment

3. **Content Delivery Service**
   - Store and serve ad assets (images, videos, HTML5)
   - Transcode video to device-appropriate formats
   - Pre-cache assets to devices ahead of schedule

4. **Analytics & Reporting**
   - Impressions tracking (per device, per ad, per venue)
   - Play confirmation events from devices
   - Aggregated reporting for advertisers

5. **Billing & Payments**
   - Stripe integration for advertiser billing
   - CPM / flat-rate / time-slot pricing models
   - Venue revenue share calculation and payout (Stripe Connect)

**Deliverables:**
- [ ] API schema design
- [ ] Device registration and heartbeat system
- [ ] Ad scheduler v1 (basic playlist rotation)
- [ ] S3 asset pipeline
- [ ] Stripe billing integration

---

### Stage 3 — Device Client (On-Screen Player)

**Goal:** Software running on venue hardware that displays ads reliably with minimal maintenance.

**Architecture:**

- **For Raspberry Pi / Linux:** Electron or Chromium in kiosk mode running a local web app
- **For Android:** Native Android app in kiosk/pinning mode
- **Protocol:** Device polls server or maintains a persistent WebSocket connection

**Core Features:**

1. **Kiosk Mode**
   - Boot directly to player on power-on
   - No user-accessible OS shell or browser chrome
   - Watchdog process restarts app on crash

2. **Playlist Sync**
   - Fetches current schedule from server on startup and at defined intervals
   - Downloads ad assets to local cache ahead of play time
   - Falls back to cached ads if network is unavailable

3. **Playback Engine**
   - Supports: JPEG/PNG (static), MP4 (video), HTML5 (animated/interactive)
   - Smooth transitions between ads
   - Configurable display duration per asset

4. **Heartbeat & Telemetry**
   - Reports online status, current playlist, last played ad every 60s
   - Sends play confirmation events (impression tracking)
   - Reports errors and device stats (CPU temp, disk space)

5. **Remote Management**
   - Accepts commands from server: reboot, update, override playlist
   - OTA software updates

6. **Offline Resilience**
   - Continues playback from local cache for up to 24h without connectivity
   - Queues impression events and flushes when reconnected

**Deliverables:**
- [ ] Linux/Pi kiosk app (Chromium-based)
- [ ] Watchdog/auto-start service
- [ ] Playlist sync and local caching
- [ ] Heartbeat and impression event reporting
- [ ] OTA update mechanism

---

### Stage 4 — Web Application (Advertiser & Venue Portal)

**Goal:** Browser-based dashboard for advertisers to buy/manage ads, and for venues to view their earnings.

**Tech Stack:**
- **Frontend:** Next.js (React) + Tailwind CSS
- **Hosting:** Vercel or Cloudflare Pages
- **Auth:** Auth.js / Clerk / Supabase Auth

**Advertiser Portal Features:**

1. **Campaign Builder**
   - Upload ad creative (image, video, HTML5)
   - Set targeting: city, venue category (restaurant, office, gym), time slots
   - Set budget, duration, CPM or flat-rate pricing
   - Preview ad as it will appear on screen

2. **Campaign Dashboard**
   - Live status of active campaigns
   - Impression counts, estimated reach, play confirmations
   - Spend tracking vs. budget

3. **Marketplace** *(see Stage 6)*

**Venue Owner Portal Features:**

1. **Venue Registration**
   - Register venue, describe location/category/foot traffic estimate
   - Request hardware (generates provisioning order)

2. **Device Status**
   - View connected devices, last heartbeat, current playlist
   - Alert if device goes offline

3. **Earnings Dashboard**
   - Revenue share from ads served
   - Payout history and upcoming payment schedule

**Admin Portal:**
- Approve new venues
- Manage all devices, campaigns, users
- Override schedules, force-push content

**Deliverables:**
- [ ] Auth and user roles (advertiser, venue, admin)
- [ ] Campaign builder and targeting UI
- [ ] Advertiser analytics dashboard
- [ ] Venue management and earnings view
- [ ] Admin panel

---

### Stage 5 — Mobile App (Venue & Advertiser Companion)

**Goal:** Mobile companion app for venue owners to monitor their display and for advertisers to check campaign performance on the go.

**Tech Stack:** React Native (Expo) — single codebase for iOS and Android.

**Features:**

**Venue Owner:**
- View device status (online/offline) at a glance
- Receive push notifications if a device goes offline
- View this month's earnings summary

**Advertiser:**
- View campaign performance (impressions, spend)
- Pause/resume campaigns
- Receive alerts when budget is nearly exhausted

**Shared:**
- Push notifications for critical events
- Biometric auth

**Out of scope (v1):** Creating campaigns, uploading creatives, venue registration — these stay in the web app.

**Deliverables:**
- [ ] Expo project scaffold with navigation
- [ ] Auth (shared with web)
- [ ] Venue device status screen
- [ ] Advertiser campaign summary screen
- [ ] Push notifications (Expo Notifications + FCM/APNs)

---

### Stage 6 — Ad Marketplace

**Goal:** A self-serve exchange where advertisers can browse available inventory by venue, time slot, and location, and purchase placements directly.

**How It Works:**
- Venues list available time slots and set floor prices
- Advertisers browse, filter, and purchase slots
- Platform takes a percentage of each transaction (15–25%)
- Programmatic fallback: unsold inventory shows house ads or partner network ads

**Marketplace Features:**

1. **Inventory Browser**
   - Map view: see available venues in a city
   - Filter by: venue category, estimated daily impressions, price range, time slot
   - Venue profile page: photos, location, foot traffic estimate, current pricing

2. **Slot Purchasing**
   - Select time slots (e.g., Tue–Thu 11am–2pm at Joe's Diner)
   - Choose duration (1 week, 1 month, etc.)
   - Upload or select creative from library
   - Checkout via Stripe

3. **Bidding (Phase 2)**
   - Auction-based pricing for premium slots
   - Reserve price set by venue

4. **Programmatic API (Phase 3)**
   - OpenRTB-compatible endpoint for DSP integration
   - Enables ad agencies and trading desks to buy inventory programmatically

**Deliverables:**
- [ ] Inventory data model and availability calendar
- [ ] Map-based and list-based inventory browser
- [ ] Slot checkout flow (Stripe)
- [ ] Venue pricing controls
- [ ] Programmatic API (phase 2)

---

### Stage 7 — Business Model & Go-to-Market

**Goal:** Define how AdPack makes money and how venues and advertisers are acquired.

**Revenue Model:**

| Source | Details |
|---|---|
| Ad Marketplace Fee | 20% of all ad spend transacted through the platform |
| Managed Campaigns | Full-service campaign management for larger advertisers (+10% markup) |
| Hardware Fee (optional) | $0 if venue commits to 12-month contract; $99 one-time otherwise |
| SaaS Subscription (Advertisers) | Optional Pro tier ($49/mo) for advanced targeting, priority support |
| Data/Analytics | Aggregated, anonymized foot-traffic reports sold to brands |

**Venue Acquisition:**
- Revenue share: venues earn **30% of net ad revenue** from their screens
- Target: restaurants, barbershops, gyms, dental offices, auto shops — anywhere with a TV and a wait
- Sales motion: direct outreach, partnerships with POS system companies, referral program
- Hardware: provide free or subsidized; recover cost over 6–12 months of ad revenue

**Advertiser Acquisition:**
- Self-serve marketplace (low friction, credit card, start in minutes)
- Target: local businesses first (restaurants, gyms, salons advertising to neighboring venues)
- Later: regional and national brands, agencies

**Partnerships:**
- POS/restaurant tech companies (Toast, Square) for venue referrals
- Local advertising agencies for managed campaign revenue
- National out-of-home networks (Lamar, Clear Channel) for remnant inventory exchange

---

### Stage 8 — Financial Estimates

**Unit Economics (Per Venue Screen):**

| Metric | Estimate |
|---|---|
| Hardware cost per device | $60–$80 |
| Monthly ad revenue per screen | $50–$200 (varies by venue traffic) |
| Platform take (20%) | $10–$40/mo per screen |
| Venue payout (30%) | $15–$60/mo per screen |
| Net revenue to AdPack per screen | $10–$40/mo |

**Pricing for Advertisers (CPM-based):**

| Tier | CPM | Notes |
|---|---|---|
| Basic (static image) | $5–$10 CPM | Lowest cost |
| Standard (video) | $10–$20 CPM | Mid-tier |
| Premium (targeted slots) | $20–$50 CPM | High-traffic venues, peak hours |

*CPM = cost per 1,000 impressions. A screen showing 4 ads/min in an 8hr/day venue = ~1,920 plays/day = ~58k plays/month.*

**Revenue Projections:**

| Milestone | Screens | Avg Net/Screen | Monthly Revenue |
|---|---|---|---|
| Pilot | 25 | $15 | $375 |
| Early Traction | 250 | $20 | $5,000 |
| Growth | 1,000 | $25 | $25,000 |
| Scale | 5,000 | $30 | $150,000 |
| Mature | 20,000 | $35 | $700,000 |

**Startup Cost Estimates:**

| Item | Cost |
|---|---|
| Pilot hardware (25 units) | $2,000 |
| Cloud infrastructure (yr 1) | $3,600 |
| Development (estimated, in-house) | — |
| Legal / business formation | $1,500 |
| Marketing / sales (yr 1) | $5,000 |
| **Total Pilot Budget** | **~$12,000** |

**Break-Even:** At ~300 active screens generating $25 net/month = $7,500/mo, which covers estimated infrastructure and basic operating costs.

---

## Staged Execution Roadmap

| Stage | Focus | Est. Duration |
|---|---|---|
| 1 | Hardware selection + pilot procurement | 2 weeks |
| 2 | Backend server core (devices, scheduling, S3) | 6–8 weeks |
| 3 | Device client (Pi kiosk, playlist sync, heartbeat) | 4–6 weeks |
| 4 | Web app (advertiser portal + venue portal) | 6–8 weeks |
| 5 | Mobile app companion | 4 weeks |
| 6 | Ad marketplace (inventory browser + checkout) | 4–6 weeks |
| 7 | Go-to-market execution | Ongoing from Stage 3 |
| 8 | Financial tracking + investor materials | Ongoing |

**Total to MVP (Stages 1–4):** ~16–22 weeks

---

## Open Questions / Decisions Needed

1. **Hardware:** Pi 5 vs. Android stick — finalize before ordering pilot batch.
2. **Pricing model:** CPM vs. time-slot flat rate vs. hybrid — impacts scheduler complexity.
3. **Revenue share %:** 30% to venues is generous; may need to adjust based on pilot data.
4. **Content moderation:** Who approves ad creatives? Automated + manual review process needed.
5. **Network connectivity:** Who pays for venue internet? Cellular LTE modem as backup?
6. **Legal:** Local regulations on digital signage vary by city/state — need legal review.
7. **DPAA / OAAA:** Consider joining Digital Place-Based Advertising Association for industry credibility.
