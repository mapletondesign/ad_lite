# AdPack — DOOH Advertising Platform

Venue screens (restaurants, offices, gyms) host a Raspberry Pi 5 running a kiosk player. Advertisers buy time slots. Venues earn 30% revenue share. AdPack takes 20%.

---

## Locked Decisions

| Decision | Choice |
|---|---|
| Backend | Go, chi v5, pgx/v5, Redis |
| Pricing v1 | Time-slot flat rate |
| Pricing v2 | CPM (after impression tracking is solid) |
| Hardware | Raspberry Pi 5 (4GB) for dev and pilot |
| Revenue share | 30% to venues (validate with pilot data) |
| Platform take | 20% of ad spend |

---

## Stages

| # | Stage | Status | Notes |
|---|---|---|---|
| 1 | Hardware selection | Done | Pi 5 locked in |
| 2 | Backend server | Done | Devices, slots, bookings, scheduler, Docker |
| 3 | Device client | Done | Vanilla HTML/JS kiosk player, heartbeat, impressions |
| 4 | Web app | Not started | Nuxt.js (Vue 3) — advertiser portal, venue portal, admin |
| 5 | Mobile app | Not started | React Native (Expo) — venue + advertiser companion |
| 6 | Ad marketplace | Not started | Self-serve inventory browser, Stripe checkout |
| 7 | Go-to-market | Ongoing from Stage 3 | Direct outreach, POS partnerships, referral program |
| 8 | Financial tracking | Ongoing | Unit economics, investor materials |

---

## Stage 4 — Web App (Nuxt.js / Vue 3)

- **Advertiser portal:** campaign builder, creative upload, targeting (city/category/time), dashboard (impressions, spend)
- **Venue portal:** registration, device status, earnings dashboard
- **Admin panel:** approve venues, manage devices/campaigns, override schedules

---

## Stage 5 — Mobile App (React Native/Expo)

- Venue: device online/offline status, push alerts, monthly earnings
- Advertiser: campaign performance, pause/resume, budget alerts
- Shared auth with web app

---

## Stage 6 — Marketplace

- Map + list inventory browser (filter by category, impressions, price, time slot)
- Slot checkout via Stripe
- Phase 2: auction-based pricing
- Phase 3: OpenRTB programmatic API for DSP integration

---

## Revenue Model

| Source | Detail |
|---|---|
| Marketplace fee | 20% of all ad spend |
| Managed campaigns | Full-service +10% markup |
| Hardware fee | $0 with 12-month contract, $99 otherwise |
| Advertiser Pro tier | $49/mo — advanced targeting, priority support |
| Data/analytics | Aggregated foot-traffic reports sold to brands |

---

## Unit Economics (per screen)

| Metric | Estimate |
|---|---|
| Hardware cost | $60–$80 |
| Monthly ad revenue | $50–$200 |
| Platform net (20%) | $10–$40/mo |
| Venue payout (30%) | $15–$60/mo |

---

## Revenue Projections

| Milestone | Screens | Net/Screen | Monthly Revenue |
|---|---|---|---|
| Pilot | 25 | $15 | $375 |
| Early traction | 250 | $20 | $5,000 |
| Growth | 1,000 | $25 | $25,000 |
| Scale | 5,000 | $30 | $150,000 |
| Mature | 20,000 | $35 | $700,000 |

---

## Pilot Budget

| Item | Cost |
|---|---|
| Hardware (25 units) | $2,000 |
| Cloud infra (yr 1) | $3,600 |
| Legal / formation | $1,500 |
| Marketing / sales | $5,000 |
| **Total** | **~$12,000** |

Break-even: ~300 screens at $25 net/mo = $7,500/mo.

---

## Open Questions

1. Pi 5 vs Android stick for production — finalize before pilot batch order
2. Content moderation — automated + manual review process needed
3. Venue internet — who pays? Cellular LTE modem as backup?
4. Legal — local digital signage regulations vary by city/state
5. DPAA/OAAA membership — consider for industry credibility
