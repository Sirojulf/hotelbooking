/**
 * Booking flow untuk RoomMasterb (Next.js / Node.js REST API)
 *
 * CARA PAKAI:
 * 1. Clone RoomMasterb: git clone https://github.com/Sirojulf/RoomMasterb.git
 * 2. Jalankan server RoomMasterb (lihat README-nya)
 * 3. Sesuaikan ROOMMASTER.baseURL di config.js
 * 4. Cari endpoint login dan reservation di source code RoomMasterb
 * 5. Update bagian "TODO" di bawah sesuai endpoint yang ada
 *
 * Asumsi awal: Next.js dengan API routes di /api/...
 * Sesuaikan path jika berbeda.
 */

import http from 'k6/http';
import { check, group } from 'k6';
import { ROOMMASTER, getCheckInOut, pickRoom } from '../config.js';

const BASE = ROOMMASTER.baseURL;
const JSON_H = { 'Content-Type': 'application/json' };
const authH = (token) => ({ ...JSON_H, Authorization: `Bearer ${token}` });

/**
 * Full booking flow untuk RoomMasterb.
 * Sesuaikan endpoint dan payload setelah kamu inspect API-nya.
 */
export function runBookingFlow() {
  const vuId = __VU - 1;
  const roomId = pickRoom(ROOMMASTER.roomIds, vuId);
  const { checkIn, checkOut } = getCheckInOut(vuId);
  const start = Date.now();

  let token = null;
  let reservationId = null;
  let success = false;

  // 1. Login
  // TODO: sesuaikan path dan field login RoomMasterb
  group('login', () => {
    const res = http.post(
      `${BASE}/auth/login`,           // ganti jika endpoint berbeda
      JSON.stringify({ email: ROOMMASTER.email, password: ROOMMASTER.password }),
      { headers: JSON_H },
    );
    if (check(res, { 'login 200': (r) => r.status === 200 })) {
      // TODO: sesuaikan field token (mungkin 'token', 'access_token', atau 'data.token')
      token = res.json('token') || res.json('access_token') || res.json('data.token');
    }
  });

  if (!token) return { success: false, durationMs: Date.now() - start };

  // 2. Cek ketersediaan
  // TODO: sesuaikan endpoint availability RoomMasterb
  group('check_availability', () => {
    const res = http.get(
      `${BASE}/rooms/${roomId}/availability?check_in=${checkIn}&check_out=${checkOut}`,
      { headers: authH(token) },
    );
    check(res, { 'availability 200': (r) => r.status === 200 });
  });

  // 3. Buat reservasi
  // TODO: sesuaikan endpoint dan field reservasi RoomMasterb
  group('create_reservation', () => {
    const res = http.post(
      `${BASE}/reservations`,         // ganti jika endpoint berbeda
      JSON.stringify({
        room_id: roomId,
        check_in_date: checkIn,       // ganti nama field jika berbeda
        check_out_date: checkOut,
        special_requests: 'k6 test',
      }),
      { headers: authH(token) },
    );
    if (check(res, { 'reservation 201': (r) => r.status === 201 || r.status === 200 })) {
      // TODO: sesuaikan field id reservasi
      reservationId = res.json('id') || res.json('data.id') || res.json('reservation.id');
    }
  });

  if (!reservationId) return { success: false, durationMs: Date.now() - start };

  // 4. Cancel reservasi (cleanup)
  // TODO: sesuaikan endpoint cancel RoomMasterb
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

/**
 * Simple read-only flow.
 * TODO: sesuaikan endpoint list hotel/room RoomMasterb
 */
export function runReadFlow() {
  const res = http.get(`${BASE}/hotels`);  // ganti jika berbeda
  const ok = check(res, { 'hotels 200': (r) => r.status === 200 });
  return { success: ok, durationMs: res.timings.duration };
}
