/**
 * Booking flow untuk hotelbooking (Go REST API)
 * Flow: login → cek availability → buat reservasi → cancel
 * Cancel di akhir agar room bisa dipakai ulang oleh VU lain
 */

import http from 'k6/http';
import { check, group } from 'k6';
import { HOTELBOOKING, getCheckInOut, pickRoom } from '../config.js';

const BASE = HOTELBOOKING.baseURL;
const JSON_H = { 'Content-Type': 'application/json' };
const authH = (token) => ({ ...JSON_H, Authorization: `Bearer ${token}` });

/**
 * Jalankan full booking flow.
 * @returns {{ success: boolean, durationMs: number }}
 */
export function runBookingFlow() {
  const vuId = __VU - 1;
  const roomId = pickRoom(HOTELBOOKING.roomIds, vuId);
  const { checkIn, checkOut } = getCheckInOut(vuId);
  const start = Date.now();

  let token = null;
  let reservationId = null;
  let success = false;

  // 1. Login
  group('login', () => {
    const res = http.post(
      `${BASE}/auth/guest/login`,
      JSON.stringify({ login: HOTELBOOKING.email, password: HOTELBOOKING.password }),
      { headers: JSON_H },
    );
    if (check(res, { 'login 200': (r) => r.status === 200 })) {
      token = res.json('access_token');
    }
  });

  if (!token) return { success: false, durationMs: Date.now() - start };

  // 2. Cek ketersediaan
  group('check_availability', () => {
    const res = http.get(
      `${BASE}/rooms/${roomId}/availability?check_in=${checkIn}&check_out=${checkOut}`,
      { headers: authH(token) },
    );
    check(res, { 'availability 200': (r) => r.status === 200 });
  });

  // 3. Buat reservasi
  group('create_reservation', () => {
    const res = http.post(
      `${BASE}/guests/reservations`,
      JSON.stringify({
        hotel_id: HOTELBOOKING.hotelId,
        room_id: roomId,
        check_in: checkIn,
        check_out: checkOut,
        booking_source: 'online',
        payment_method: 'transfer',
        special_requests: 'k6 test',
      }),
      { headers: authH(token) },
    );
    if (check(res, { 'reservation 201': (r) => r.status === 201 })) {
      reservationId = res.json('reservation.id');
    }
  });

  if (!reservationId) return { success: false, durationMs: Date.now() - start };

  // 4. Cancel (cleanup — biarkan slot terbuka untuk VU lain)
  group('cancel_reservation', () => {
    const res = http.post(
      `${BASE}/guests/reservations/${reservationId}/cancel`,
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
 * Simple read-only flow: list hotels.
 * Digunakan saat mengukur throughput murni tanpa autentikasi.
 */
export function runReadFlow() {
  const res = http.get(`${BASE}/hotels`);
  const ok = check(res, { 'hotels 200': (r) => r.status === 200 });
  return { success: ok, durationMs: res.timings.duration };
}
