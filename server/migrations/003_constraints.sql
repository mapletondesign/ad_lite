-- Migration 003 — data integrity CHECK constraints

-- bookings
ALTER TABLE bookings
  ADD CONSTRAINT chk_bookings_date_order  CHECK (ends_on >= starts_on),
  ADD CONSTRAINT chk_bookings_price_cents CHECK (price_cents > 0),
  ADD CONSTRAINT chk_bookings_status      CHECK (status IN ('pending', 'active', 'completed', 'cancelled'));

-- ad_slots
ALTER TABLE ad_slots
  ADD CONSTRAINT chk_ad_slots_time_order   CHECK (end_time > start_time),
  ADD CONSTRAINT chk_ad_slots_price_cents  CHECK (price_cents > 0),
  ADD CONSTRAINT chk_ad_slots_duration_sec CHECK (duration_sec > 0),
  ADD CONSTRAINT chk_ad_slots_status       CHECK (status IN ('available', 'booked', 'paused')),
  ADD CONSTRAINT chk_ad_slots_days_nonempty CHECK (array_length(days_of_week, 1) > 0),
  ADD CONSTRAINT chk_ad_slots_days_range    CHECK (days_of_week <@ ARRAY[0,1,2,3,4,5,6]::int[]);
