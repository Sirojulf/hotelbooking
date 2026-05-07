/**
 * Booking flow untuk RoomMasterb (Node.js REST API)
 * Flow: cek availability → buat reservasi → cancel
 * Token login di-cache via setup() agar tidak rate-limited.
 */

import http from 'k6/http';
import { check, group } from 'k6';
import { ROOMMASTER, getCheckInOut, pickRoom } from '../config.js';

const BASE   = ROOMMASTER.baseURL;
const JSON_H = { 'Content-Type': 'application/json' };
const authH  = (token) => ({ ...JSON_H, Authorization: `Bearer ${token}` });

/**
 * Login sekali sebelum semua VU mulai.
 * Hasilnya dikirim ke setiap VU melalui parameter fungsi default.
 */
export function setup() {
  const res = http.post(
    `${BASE}/auth/login`,
    JSON.stringify({ email: ROOMMASTER.email, password: ROOMMASTER.password }),
    { headers: JSON_H },
  );
  if (res.status !== 200) {
    console.error(`[setup] Login gagal: ${res.status} - ${res.body}`);
    return { token: null };
  }
  const token = res.json('access_token');
  if (!token) {
    console.error('[setup] access_token tidak ditemukan di response');
    return { token: null };
  }
  console.log('[setup] Login berhasil, token siap dipakai semua VU');
  return { token };
}

/**
 * Jalankan full booking flow (tanpa login ulang).
 * @param {string} token  - JWT dari setup()
 * @returns {{ success: boolean, durationMs: number }}
 */
export function runBookingFlow(token) {
  const vuId   = __VU - 1;
  const iter   = __ITER;
  const roomId = pickRoom(ROOMMASTER.roomIds, vuId);
  const { checkIn, checkOut } = getCheckInOut(vuId, iter);
  const start  = Date.now();

  if (!token) {
    console.error(`VU${__VU} iter${iter}: token null, skip`);
    return { success: false, durationMs: 0 };
  }

  let reservationId = null;
  let success = false;

  // 1. Cek ketersediaan
  group('check_availability', () => {
    const res = http.get(
      `${BASE}/rooms/${roomId}/availability?check_in=${checkIn}&check_out=${checkOut}`,
      { headers: authH(token) },
    );
    check(res, { 'availability 200': (r) => r.status === 200 });
  });

  // 2. Buat reservasi
  group('create_reservation', () => {
    const res = http.post(
      `${BASE}/reservations`,
      JSON.stringify({
        room_id:          roomId,
        check_in:         checkIn,
        check_out:        checkOut,
        special_requests: 'k6 test',
      }),
      { headers: authH(token) },
    );
    if (check(res, { 'reservation 201': (r) => r.status === 201 || r.status === 200 })) {
      reservationId = res.json('reservation.id');
    }
  });

  if (!reservationId) return { success: false, durationMs: Date.now() - start };

  // 3. Cancel — bebaskan slot untuk iterasi berikutnya
  group('cancel_reservation', () => {
    const res = http.post(
      `${BASE}/reservations/${reservationId}/cancel`,
      null,
      { headers: authH(token) },
    );
    if (check(res, { 'cancel 200': (r) => r.status === 200 })) {
      success = true;
    }
  });

  return { success, durationMs: Date.now() - start };
}

// Entry point untuk k6 run langsung (quick test)
export default function (data) {
  runBookingFlow(data ? data.token : null);
}

export function runReadFlow() {
  const res = http.get(`${BASE}/hotels`);
  const ok  = check(res, { 'hotels 200': (r) => r.status === 200 });
  return { success: ok, durationMs: res.timings.duration };
}
