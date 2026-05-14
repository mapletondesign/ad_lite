export interface TokenPair {
  access_token: string
  refresh_token: string
  expires_in: number
}

export interface AuthUser {
  id: string
  role: 'admin' | 'advertiser' | 'venue'
  advertiser_id?: string
  venue_id?: string
}

export interface Venue {
  id: string
  name: string
  category: string | null
  address: string | null
  city: string | null
  state: string | null
  country: string
  created_at: string
  updated_at: string
}

export interface Advertiser {
  id: string
  name: string
  email: string
  created_at: string
}

export interface Device {
  id: string
  venue_id: string
  name: string
  status: 'online' | 'offline'
  last_seen: string | null
  ip_address: string | null
  firmware_version: string | null
  created_at: string
}

export interface AdSlot {
  id: string
  device_id: string
  label: string | null
  days_of_week: number[]
  start_time: string
  end_time: string
  duration_sec: number
  price_cents: number
  status: 'available' | 'booked' | 'paused'
  created_at: string
}

export interface Booking {
  id: string
  slot_id: string
  advertiser_id: string
  creative_url: string | null
  starts_on: string
  ends_on: string
  price_cents: number
  status: 'pending' | 'active' | 'completed' | 'cancelled'
  created_at: string
}
