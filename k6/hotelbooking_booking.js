/**
 * k6 HTTP test — hotelbooking REST API
 * Booking flow: login → cek availability → buat reservasi → bayar → cancel
 *
 * Jalankan:
 *   k6 run k6/hotelbooking_booking.js
 *   k6 run --vus 10 --duration 1m k6/hotelbooking_booking.js
 */

import http from 'k6/http';
import { check, group, sleep } from 'k6';
import { Rate, Trend, Counter } from 'k6/metrics';
import { HOTELBOOKING, getCheckInOut, pickRoom } from './config.js';

// ─── Custom metrics ───────────────────────────────────────────────────────────
const bookingSuccessRate  = new Rate('booking_success_rate');
const bookingFlowDuration = new Trend('booking_flow_duration_ms', true);
const bookingTotal        = new Counter('booking_total');
const paymentTotal        = new Counter('payment_total');

// ─── Load test config ─────────────────────────────────────────────────────────
export const options = {
  scenarios: {
    booking_ramp: {
      executor: 'ramping-vus',
      startVUs: 1,
      stages: [
        { duration: '30s', target: 5  },  // ramp up
        { duration: '1m',  target: 5  },  // steady state
        { duration: '30s', target: 10 },  // spike
        { duration: '30s', target: 0  },  // ramp down
      ],
    },
  },
  thresholds: {
    // 95% request harus selesai < 2 detik
    http_req_duration:     ['p(95)<2000'],
    // Minimal 90% booking berhasil
    booking_success_rate:  ['rate>0.9'],
    // Flow lengkap < 5 detik di P95
    booking_flow_duration_ms: ['p(95)<5000'],
  },
};

const BASE = HOTELBOOKING.baseURL;
const JSON_HEADERS = { 'Content-Type': 'application/json' };

function authHeaders(token) {
  return { ...JSON_HEADERS, Authorization: `Bearer ${token}` };
}

// ─── Main test function ───────────────────────────────────────────────────────
export default function () {
  const vuId      = __VU - 1;
  const roomId    = pickRoom(HOTELBOOKING.roomIds, vuId);
  const { checkIn, checkOut } = getCheckInOut(vuId);

  const flowStart  = Date.now();
  let token        = null;
  let reservationId = null;
  let flowSuccess  = false;

  // ── 1. Login ────────────────────────────────────────────────────────────────
  group('1_login', () => {
    const res = http.post(
      `${BASE}/auth/guest/login`,
      JSON.stringify({ login: HOTELBOOKING.email, password: HOTELBOOKING.password }),
      { headers: JSON_HEADERS },
    );

    const ok = check(res, {
      'login: status 200':       (r) => r.status === 200,
      'login: ada access_token': (r) => !!r.json('access_token'),
    });

    if (ok) token = res.json('access_token');
  });

  if (!token) {
    bookingSuccessRate.add(false);
    return;
  }

  // ── 2. Cek ketersediaan kamar ────────────────────────────────────────────────
  group('2_check_availability', () => {
    const res = http.get(
      `${BASE}/rooms/${roomId}/availability?check_in=${checkIn}&check_out=${checkOut}`,
      { headers: authHeaders(token) },
    );

    check(res, {
      'availability: status 200':  (r) => r.status === 200,
      'availability: ada field available': (r) => r.json('available') !== undefined,
    });
    // Tidak stop jika tidak available — tetap coba buat reservasi
    // agar kita bisa mengukur error handling-nya
  });

  // ── 3. Buat reservasi ────────────────────────────────────────────────────────
  group('3_create_reservation', () => {
    const res = http.post(
      `${BASE}/guests/reservations`,
      JSON.stringify({
        hotel_id:         HOTELBOOKING.hotelId,
        room_id:          roomId,
        check_in:         checkIn,
        check_out:        checkOut,
        booking_source:   'online',
        payment_method:   'transfer',
        special_requests: 'k6 load test',
      }),
      { headers: authHeaders(token) },
    );

    const ok = check(res, {
      'create reservation: status 201':     (r) => r.status === 201,
      'create reservation: ada id':         (r) => !!r.json('reservation.id'),
      'create reservation: status pending': (r) => r.json('reservation.payment_status') === 'pending',
    });

    if (ok) {
      reservationId = res.json('reservation.id');
      bookingTotal.add(1);
    }
  });

  if (!reservationId) {
    bookingSuccessRate.add(false);
    bookingFlowDuration.add(Date.now() - flowStart);
    return;
  }

  // ── 4. Bayar reservasi ───────────────────────────────────────────────────────
  group('4_pay_reservation', () => {
    const res = http.post(
      `${BASE}/guests/reservations/${reservationId}/pay`,
      JSON.stringify({ payment_method: 'transfer' }),
      { headers: authHeaders(token) },
    );

    const ok = check(res, {
      'pay: status 200':    (r) => r.status === 200,
      'pay: status = paid': (r) => r.json('reservation.payment_status') === 'paid',
    });

    if (ok) {
      flowSuccess = true;
      paymentTotal.add(1);
    }
  });

  // ── 5. Cancel reservasi (cleanup agar room bisa dipakai VU lain) ─────────────
  group('5_cancel_reservation', () => {
    const res = http.post(
      `${BASE}/guests/reservations/${reservationId}/cancel`,
      null,
      { headers: authHeaders(token) },
    );

    check(res, {
      'cancel: status 200': (r) => r.status === 200,
    });
  });

  // ── Record metrics ───────────────────────────────────────────────────────────
  bookingSuccessRate.add(flowSuccess);
  bookingFlowDuration.add(Date.now() - flowStart);

  sleep(1);
}

// ─── Summary di akhir test ────────────────────────────────────────────────────
export function handleSummary(data) {
  const metrics = data.metrics;

  const p95Duration = metrics.booking_flow_duration_ms?.values?.['p(95)'] ?? 0;
  const successRate = (metrics.booking_success_rate?.values?.rate ?? 0) * 100;
  const totalBookings = metrics.booking_total?.values?.count ?? 0;
  const reqDuration = metrics.http_req_duration?.values?.['p(95)'] ?? 0;

  console.log('\n╔══════════════════════════════════════════╗');
  console.log('║   HOTELBOOKING — HASIL BOOKING TEST      ║');
  console.log('╠══════════════════════════════════════════╣');
  console.log(`║ Booking sukses rate : ${successRate.toFixed(1).padStart(6)}%             ║`);
  console.log(`║ Total booking       : ${String(totalBookings).padStart(6)} kali           ║`);
  console.log(`║ Flow duration P95   : ${p95Duration.toFixed(0).padStart(6)} ms            ║`);
  console.log(`║ HTTP req P95        : ${reqDuration.toFixed(0).padStart(6)} ms            ║`);
  console.log('╚══════════════════════════════════════════╝\n');

  return {
    'k6/results/hotelbooking_summary.json': JSON.stringify(data, null, 2),
  };
}
