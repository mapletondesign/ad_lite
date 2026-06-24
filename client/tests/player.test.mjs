/**
 * Tests for client/js/player.js resilience behaviours.
 *
 * Uses Node 24's built-in test runner (node:test) and a lightweight VM
 * sandbox — no browser or external dependencies required.
 *
 * Timing constants (RETRY_DELAY, NO_ADS_DELAY) are patched to 0 so the
 * player loop cycles through in a single event-loop turn instead of waiting
 * 10–30 seconds.  HEARTBEAT_INTERVAL is set to a large valid value so the
 * independent setInterval doesn't fire during tests.
 */

import { test } from 'node:test';
import assert from 'node:assert/strict';
import { readFileSync } from 'node:fs';
import vm from 'node:vm';
import { fileURLToPath } from 'node:url';
import path from 'node:path';

// ── Helpers ───────────────────────────────────────────────────────────────────

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const RAW_SRC = readFileSync(path.join(__dirname, '../js/player.js'), 'utf8');

function patchSrc(src) {
  return src
    .replace(/const RETRY_DELAY\s*=\s*\S+/,        'const RETRY_DELAY = 0')
    .replace(/const NO_ADS_DELAY\s*=\s*\S+/,       'const NO_ADS_DELAY = 0')
    // 2 000 000 ms (~33 min) — large but within int32 range so no TimeoutOverflowWarning.
    .replace(/const HEARTBEAT_INTERVAL\s*=\s*\S+/,  'const HEARTBEAT_INTERVAL = 2_000_000');
}

// Returns a minimal DOM element whose classList.add calls are recorded in `history`.
function makeElement(history = null) {
  const classes = new Set();
  return {
    title: '',
    src:   '',
    onload: null, onerror: null, onended: null,
    play()  { return Promise.resolve(); },
    pause() {},
    addEventListener() {},
    classList: {
      add(c)      { if (history) history.push(c); classes.add(c); },
      remove(c)   { classes.delete(c); },
      contains(c) { return classes.has(c); },
    },
    get className()  { return [...classes].join(' '); },
    set className(v) {
      classes.clear();
      if (v) v.split(/\s+/).filter(Boolean).forEach(c => classes.add(c));
    },
  };
}

// Builds a vm sandbox for one test run.
//  - deviceId   : the device_id URL param value
//  - fetchImpl  : async function(url, opts) replacing global fetch
//  - dotHistory : array that records every setStatus() value applied to status-dot
function makeContext(deviceId, fetchImpl, dotHistory) {
  const dot = makeElement(dotHistory);
  const stub = makeElement();

  const storage = new Map([['adlite_device_token', 'test-token']]);

  const sandbox = {
    URLSearchParams,
    location:     { search: `?device_id=${deviceId}&token=test-token` },
    localStorage: { getItem: (k) => storage.get(k) ?? null, setItem: (k, v) => storage.set(k, v) },
    fetch:        fetchImpl,
    document: {
      getElementById(id) { return id === 'status-dot' ? dot : stub; },
    },
    // JS built-ins the player uses
    Array, Math, Date, Promise, JSON, parseInt, parseFloat,
    isNaN, isFinite, console, Error, TypeError, RegExp,
    Object, Function, Symbol, Map, Set,
  };

  // Live timer getters so any future mocking of globalThis propagates.
  for (const name of ['setTimeout', 'setInterval', 'clearTimeout']) {
    Object.defineProperty(sandbox, name, { get: () => globalThis[name], configurable: true });
  }

  return vm.createContext(sandbox);
}

// Yields to the event loop several times so async loop iterations can complete.
async function flush(rounds = 8) {
  for (let i = 0; i < rounds; i++) {
    await new Promise(r => setImmediate(r));
  }
}

// ── Tests ─────────────────────────────────────────────────────────────────────

test('empty playlist: shows "waiting" status and keeps polling', async () => {
  let callCount = 0;
  const statusHistory = [];

  const ctx = makeContext('device-abc', async () => {
    callCount++;
    return { ok: true, json: async () => ({ playlist: [] }) };
  }, statusHistory);

  vm.runInContext(patchSrc(RAW_SRC), ctx);

  // Let the loop complete several iterations (NO_ADS_DELAY = 0 so it cycles fast).
  await flush();

  assert.ok(
    statusHistory.includes('waiting'),
    `"waiting" status was never set; status history: [${statusHistory}]`,
  );
  assert.ok(
    callCount >= 2,
    `heartbeat should have been called more than once (got ${callCount}), proving the loop re-polls after an empty playlist`,
  );
});

test('unreachable server: shows "error" status and retries', async () => {
  let callCount = 0;
  const statusHistory = [];

  const ctx = makeContext('device-xyz', async () => {
    callCount++;
    throw new TypeError('Failed to fetch');
  }, statusHistory);

  vm.runInContext(patchSrc(RAW_SRC), ctx);

  // Let the loop attempt, fail, and retry at least once.
  await flush();

  assert.ok(
    statusHistory.includes('error'),
    `"error" status was never set after a network failure; status history: [${statusHistory}]`,
  );
  assert.ok(
    callCount >= 2,
    `heartbeat should have been retried (got ${callCount} attempt(s)), proving the loop recovers from network errors`,
  );
});
