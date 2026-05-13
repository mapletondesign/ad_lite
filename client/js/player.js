'use strict';

// ── Config ────────────────────────────────────────────────────────────────────

const STORAGE_KEY        = 'adlite_device_id';
const HEARTBEAT_INTERVAL = 60 * 1000;  // keep server status updated
const RETRY_DELAY        = 10 * 1000;  // wait before retrying on error
const NO_ADS_DELAY       = 30 * 1000;  // poll interval when playlist is empty

// ── Elements ──────────────────────────────────────────────────────────────────

const $player  = document.getElementById('player');
const $setup   = document.getElementById('setup');
const $image   = document.getElementById('ad-image');
const $video   = document.getElementById('ad-video');
const $dot     = document.getElementById('status-dot');
const $input   = document.getElementById('device-id-input');
const $btn     = document.getElementById('setup-btn');

// ── State ─────────────────────────────────────────────────────────────────────

let deviceId = new URLSearchParams(location.search).get('device_id')
             || localStorage.getItem(STORAGE_KEY);

// ── Setup screen ──────────────────────────────────────────────────────────────

if (deviceId) {
  startPlayer(deviceId);
} else {
  $setup.classList.remove('hidden');
}

$btn.addEventListener('click', () => {
  const id = $input.value.trim();
  if (!id) return;
  localStorage.setItem(STORAGE_KEY, id);
  $setup.classList.add('hidden');
  startPlayer(id);
});

$input.addEventListener('keydown', (e) => {
  if (e.key === 'Enter') $btn.click();
});

// ── Player ────────────────────────────────────────────────────────────────────

function startPlayer(id) {
  deviceId = id;
  $player.classList.remove('hidden');
  setStatus('waiting');
  playerLoop();
  // Independent heartbeat keeps the server's last_seen timestamp fresh
  // even when the playlist is long and re-fetches are infrequent.
  setInterval(() => heartbeat().catch(() => {}), HEARTBEAT_INTERVAL);
}

async function playerLoop() {
  let playlist = [];

  while (true) {
    if (playlist.length === 0) {
      setStatus('waiting');
      try {
        playlist = await heartbeat();
      } catch (err) {
        setStatus('error');
        await sleep(RETRY_DELAY);
        continue;
      }
    }

    if (playlist.length === 0) {
      await sleep(NO_ADS_DELAY);
      continue;
    }

    const ad = playlist.shift();
    setStatus('playing');

    await playAd(ad).catch(() => {}); // never let a single ad crash the loop
    recordImpression(ad.booking_id);  // fire-and-forget
  }
}

// ── API calls ─────────────────────────────────────────────────────────────────

async function heartbeat() {
  const res = await fetch(`/api/v1/devices/${deviceId}/heartbeat`, {
    method:  'POST',
    headers: { 'Content-Type': 'application/json' },
    body:    JSON.stringify({
      ip_address:       '',
      firmware_version: 'web-client-1.0',
    }),
  });
  if (!res.ok) throw new Error(`heartbeat ${res.status}`);
  const data = await res.json();
  return Array.isArray(data.playlist) ? data.playlist : [];
}

function recordImpression(bookingId) {
  fetch(`/api/v1/devices/${deviceId}/impression`, {
    method:  'POST',
    headers: { 'Content-Type': 'application/json' },
    body:    JSON.stringify({
      booking_id: bookingId,
      played_at:  Math.floor(Date.now() / 1000),
    }),
  }).catch(() => {}); // non-blocking — player must never stall on analytics
}

// ── Playback ──────────────────────────────────────────────────────────────────

const VIDEO_EXTS = /\.(mp4|webm|ogg|mov)(\?.*)?$/i;

function playAd(ad) {
  return VIDEO_EXTS.test(ad.creative_url)
    ? playVideo(ad)
    : playImage(ad);
}

function playImage(ad) {
  return new Promise((resolve) => {
    $video.classList.remove('active');
    $image.onload = () => {
      $image.classList.add('active');
      setTimeout(() => {
        $image.classList.remove('active');
        setTimeout(resolve, 300); // wait for fade-out
      }, ad.duration_sec * 1000);
    };
    $image.onerror = resolve; // skip broken asset
    $image.src = ad.creative_url;
  });
}

function playVideo(ad) {
  return new Promise((resolve) => {
    $image.classList.remove('active');
    const done = () => {
      clearTimeout(timer);
      $video.classList.remove('active');
      $video.pause();
      $video.src = '';
      setTimeout(resolve, 300);
    };
    const timer = setTimeout(done, ad.duration_sec * 1000);
    $video.onended = done;
    $video.onerror = done;
    $video.src = ad.creative_url;
    $video.classList.add('active');
    $video.play().catch(done);
  });
}

// ── Helpers ───────────────────────────────────────────────────────────────────

function setStatus(state) {
  $dot.className   = '';
  $dot.title       = state;
  $dot.classList.add(state);
}

function sleep(ms) {
  return new Promise((r) => setTimeout(r, ms));
}
