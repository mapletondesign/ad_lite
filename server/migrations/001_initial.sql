-- Migration 001 — initial schema

CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- Venues: physical locations that host devices
CREATE TABLE venues (
  id          uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  name        text NOT NULL,
  category    text,                          -- 'restaurant' | 'gym' | 'office' | 'retail'
  address     text,
  city        text,
  state       text,
  country     text NOT NULL DEFAULT 'US',
  created_at  timestamptz NOT NULL DEFAULT now(),
  updated_at  timestamptz NOT NULL DEFAULT now()
);

-- Devices: media players installed at venues
CREATE TABLE devices (
  id               uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  venue_id         uuid NOT NULL REFERENCES venues(id) ON DELETE CASCADE,
  name             text NOT NULL,
  status           text NOT NULL DEFAULT 'offline',  -- 'online' | 'offline'
  last_seen        timestamptz,
  ip_address       text,
  firmware_version text,
  created_at       timestamptz NOT NULL DEFAULT now(),
  updated_at       timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX idx_devices_venue_id ON devices(venue_id);
CREATE INDEX idx_devices_status   ON devices(status);

-- Ad slots: time windows on a device that can be sold
-- Time-slot flat rate model (v1)
CREATE TABLE ad_slots (
  id           uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  device_id    uuid NOT NULL REFERENCES devices(id) ON DELETE CASCADE,
  label        text,                          -- e.g. 'Lunch Rush', 'Morning'
  days_of_week int[] NOT NULL,               -- 0=Sun … 6=Sat
  start_time   time NOT NULL,
  end_time     time NOT NULL,
  duration_sec int NOT NULL DEFAULT 15,      -- length of one ad play in seconds
  price_cents  int NOT NULL,
  status       text NOT NULL DEFAULT 'available', -- 'available' | 'booked' | 'paused'
  created_at   timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX idx_ad_slots_device_id ON ad_slots(device_id);

-- Advertisers: companies or individuals buying ad slots
CREATE TABLE advertisers (
  id         uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  name       text NOT NULL,
  email      text NOT NULL UNIQUE,
  created_at timestamptz NOT NULL DEFAULT now()
);

-- Bookings: an advertiser purchases a slot for a date range
CREATE TABLE bookings (
  id            uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  slot_id       uuid NOT NULL REFERENCES ad_slots(id),
  advertiser_id uuid NOT NULL REFERENCES advertisers(id),
  creative_url  text,                         -- S3 URL of uploaded ad asset
  starts_on     date NOT NULL,
  ends_on       date NOT NULL,
  price_cents   int NOT NULL,
  status        text NOT NULL DEFAULT 'pending', -- 'pending' | 'active' | 'completed' | 'cancelled'
  created_at    timestamptz NOT NULL DEFAULT now(),
  updated_at    timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX idx_bookings_slot_id       ON bookings(slot_id);
CREATE INDEX idx_bookings_advertiser_id ON bookings(advertiser_id);
CREATE INDEX idx_bookings_status        ON bookings(status);

-- Impressions: each confirmed play event reported by a device
CREATE TABLE impressions (
  id         bigserial PRIMARY KEY,
  booking_id uuid NOT NULL REFERENCES bookings(id),
  device_id  uuid NOT NULL REFERENCES devices(id),
  played_at  timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX idx_impressions_booking_id ON impressions(booking_id);
CREATE INDEX idx_impressions_device_id  ON impressions(device_id);
CREATE INDEX idx_impressions_played_at  ON impressions(played_at DESC);
