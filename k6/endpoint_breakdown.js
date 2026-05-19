/**
 * ENDPOINT BREAKDOWN — ukur durasi tiap endpoint di booking flow secara terpisah.
 * Tujuan: identifikasi endpoint mana yang paling berat (untuk paper).
 *
 * Per iterasi, satu VU jalankan urutan:
 *   1. GET  /rooms/{id}/availability
 *   2. POST /guests/reservations          (create)
 *   3. POST /guests/reservations/{id}/pay
 *   4. POST /guests/reservations/{id}/cancel
 *
 * Tiap endpoint punya Trend metric sendiri → bisa dilihat avg/med/p95/p99 per endpoint.
 *
 * Jalankan:
 *   k6 run k6/endpoint_breakdown.js
 *   k6 run -e TARGET=hotelbooking k6/endpoint_breakdown.js   # default
 *   k6 run -e TARGET=roommaster   k6/endpoint_breakdown.js   # untuk Node.js
 */
import http from 'k6/http';
import { check, sleep } from 'k6';
import { Trend, Rate } from 'k6/metrics';
import { HOTELBOOKING, ROOMMASTER, getCheckInOut, pickRoom } from './config.js';

const TARGET = __ENV.TARGET || 'hotelbooking';
const LABEL  = __ENV.LABEL  || TARGET;   // pembeda nama file output (hotelbooking/purehttp/roommaster)
const cfg = TARGET === 'roommaster' ? ROOMMASTER : HOTELBOOKING;

// ─── Per-endpoint Trend metrics ──────────────────────────────────────────────
const tAvail  = new Trend('ep_check_availability_ms', true);
const tCreate = new Trend('ep_create_reservation_ms', true);
const tPay    = new Trend('ep_pay_reservation_ms', true);
const tCancel = new Trend('ep_cancel_reservation_ms', true);

const rAvail  = new Rate('ep_check_availability_ok');
const rCreate = new Rate('ep_create_reservation_ok');
const rPay    = new Rate('ep_pay_reservation_ok');
const rCancel = new Rate('ep_cancel_reservation_ok');

export const options = {
  summaryTrendStats: ['avg', 'min', 'med', 'p(90)', 'p(95)', 'p(99)', 'max'],
  scenarios: {
    breakdown: {
      executor: 'constant-vus',
      vus: 10,
      duration: '2m',
    },
  },
};

// ─── Setup: login sekali, share token ke semua VU ────────────────────────────
export function setup() {
  const isHB = TARGET === 'hotelbooking';
  const loginURL = isHB ? `${cfg.baseURL}/auth/guest/login` : `${cfg.baseURL}/auth/login`;
  const loginBody = isHB
    ? { login: cfg.email, password: cfg.password }
    : { email: cfg.email, password: cfg.password };

  const res = http.post(loginURL, JSON.stringify(loginBody), {
    headers: { 'Content-Type': 'application/json' },
  });
  if (res.status !== 200) {
    console.error(`[setup] Login gagal: ${res.status}`);
    return { token: null };
  }
  return { token: res.json('access_token') };
}

export default function (data) {
  if (!data.token) return;

  const vuId = __VU - 1;
  const iter = __ITER;
  const roomId = pickRoom(cfg.roomIds, vuId);
  const { checkIn, checkOut } = getCheckInOut(vuId, iter);
  const headers = {
    'Content-Type': 'application/json',
    Authorization: `Bearer ${data.token}`,
  };

  // 1. CHECK AVAILABILITY ─────────────────────────────────────────────────────
  let t0 = Date.now();
  let res = http.get(
    `${cfg.baseURL}/rooms/${roomId}/availability?check_in=${checkIn}&check_out=${checkOut}`,
    { headers, tags: { endpoint: 'check_availability' } },
  );
  tAvail.add(Date.now() - t0);
  rAvail.add(res.status === 200);

  // 2. CREATE RESERVATION ─────────────────────────────────────────────────────
  t0 = Date.now();
  res = http.post(
    `${cfg.baseURL}/guests/reservations`,
    JSON.stringify({
      hotel_id: cfg.hotelId || undefined,
      room_id: roomId,
      check_in: checkIn,
      check_out: checkOut,
      booking_source: 'online',
      payment_method: 'transfer',
      special_requests: 'endpoint breakdown',
    }),
    { headers, tags: { endpoint: 'create_reservation' } },
  );
  tCreate.add(Date.now() - t0);
  const createOK = res.status === 201 || res.status === 200;
  rCreate.add(createOK);

  if (!createOK) {
    sleep(1);
    return;
  }
  const reservationId = res.json('reservation.id');
  if (!reservationId) {
    sleep(1);
    return;
  }

  // 3. PAY RESERVATION ────────────────────────────────────────────────────────
  t0 = Date.now();
  res = http.post(
    `${cfg.baseURL}/guests/reservations/${reservationId}/pay`,
    JSON.stringify({ payment_method: 'transfer' }),
    { headers, tags: { endpoint: 'pay_reservation' } },
  );
  tPay.add(Date.now() - t0);
  rPay.add(res.status === 200);

  // 4. CANCEL RESERVATION ─────────────────────────────────────────────────────
  t0 = Date.now();
  res = http.post(
    `${cfg.baseURL}/guests/reservations/${reservationId}/cancel`,
    null,
    { headers, tags: { endpoint: 'cancel_reservation' } },
  );
  tCancel.add(Date.now() - t0);
  rCancel.add(res.status === 200);

  sleep(1);
}

// ─── Summary di akhir test ───────────────────────────────────────────────────
export function handleSummary(data) {
  const m = data.metrics;

  function row(label, trendKey, rateKey) {
    const t = m[trendKey]?.values || {};
    const r = m[rateKey]?.values?.rate ?? 0;
    return {
      label,
      avg: Math.round(t.avg ?? 0),
      med: Math.round(t.med ?? 0),
      p95: Math.round(t['p(95)'] ?? 0),
      p99: Math.round(t['p(99)'] ?? 0),
      max: Math.round(t.max ?? 0),
      ok: (r * 100).toFixed(1),
    };
  }

  const rows = [
    row('check_availability', 'ep_check_availability_ms', 'ep_check_availability_ok'),
    row('create_reservation', 'ep_create_reservation_ms', 'ep_create_reservation_ok'),
    row('pay_reservation',    'ep_pay_reservation_ms',    'ep_pay_reservation_ok'),
    row('cancel_reservation', 'ep_cancel_reservation_ms', 'ep_cancel_reservation_ok'),
  ];

  console.log('\n╔════════════════════════════════════════════════════════════════════╗');
  console.log(`║   ENDPOINT BREAKDOWN — Label: ${LABEL.padEnd(35)}║`);
  console.log('╠════════════════════════════════════════════════════════════════════╣');
  console.log('║ Endpoint              Avg     Med    P95    P99    Max     OK ║');
  console.log('╠════════════════════════════════════════════════════════════════════╣');
  for (const r of rows) {
    console.log(
      `║ ${r.label.padEnd(20)}` +
      ` ${String(r.avg).padStart(4)}ms` +
      ` ${String(r.med).padStart(4)}ms` +
      ` ${String(r.p95).padStart(4)}ms` +
      ` ${String(r.p99).padStart(4)}ms` +
      ` ${String(r.max).padStart(4)}ms` +
      ` ${r.ok.padStart(5)}% ║`
    );
  }
  console.log('╚════════════════════════════════════════════════════════════════════╝\n');

  return {
    [`k6/results/endpoint_breakdown_${LABEL}.json`]: JSON.stringify(data, null, 2),
  };
}
