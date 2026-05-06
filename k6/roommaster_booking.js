/**
 * k6 Browser test — RoomMasterb (Next.js)
 * Booking flow: login sebagai FO → buka halaman reservasi → buat reservasi baru
 *
 * Jalankan:
 *   k6 run k6/roommaster_booking.js
 *   k6 run --env HEADLESS=false k6/roommaster_booking.js   ← lihat browser
 *
 * Catatan:
 *   Browser test tidak bisa high concurrency seperti HTTP test.
 *   VU=2-3 sudah cukup untuk mengukur performa UI/SSR.
 */

import { browser } from 'k6/browser';
import { check, sleep } from 'k6';
import { Rate, Trend, Counter } from 'k6/metrics';
import { ROOMMASTER, getCheckInOut, pickRoom } from './config.js';

// ─── Custom metrics ───────────────────────────────────────────────────────────
const bookingSuccessRate  = new Rate('booking_success_rate');
const bookingFlowDuration = new Trend('booking_flow_duration_ms', true);
const pageLoadDuration    = new Trend('page_load_duration_ms', true);
const bookingTotal        = new Counter('booking_total');

// ─── Load test config ─────────────────────────────────────────────────────────
export const options = {
  scenarios: {
    booking_browser: {
      executor: 'constant-vus',
      vus: 2,          // browser test: pakai VU sedikit, lebih realistis
      duration: '2m',
      options: {
        browser: { type: 'chromium' },
      },
    },
  },
  thresholds: {
    booking_success_rate:     ['rate>0.8'],
    booking_flow_duration_ms: ['p(95)<15000'],  // SSR lebih lambat dari pure API
    page_load_duration_ms:    ['p(95)<5000'],
  },
};

const BASE = ROOMMASTER.baseURL;

// ─── Helper: tunggu elemen muncul ─────────────────────────────────────────────
async function waitAndClick(page, selector, timeout = 10000) {
  await page.waitForSelector(selector, { timeout });
  await page.locator(selector).click();
}

async function waitAndFill(page, selector, value, timeout = 10000) {
  await page.waitForSelector(selector, { timeout });
  await page.locator(selector).fill(value);
}

// ─── Main test function ───────────────────────────────────────────────────────
export default async function () {
  const vuId   = __VU - 1;
  const roomId = pickRoom(ROOMMASTER.roomIds, vuId);
  const { checkIn, checkOut } = getCheckInOut(vuId);

  const flowStart = Date.now();
  let flowSuccess = false;

  const page = await browser.newPage();

  // Supaya lebih realistis — viewport layar laptop FO
  await page.setViewportSize({ width: 1280, height: 800 });

  try {
    // ── 1. Login ──────────────────────────────────────────────────────────────
    const loginStart = Date.now();
    await page.goto(`${BASE}/auth/login`, { waitUntil: 'networkidle' });
    pageLoadDuration.add(Date.now() - loginStart);

    await waitAndFill(page, 'input[name="email"], input[type="email"]', ROOMMASTER.email);
    await waitAndFill(page, 'input[name="password"], input[type="password"]', ROOMMASTER.password);
    await waitAndClick(page, 'button[type="submit"]');

    // Tunggu redirect ke dashboard FO
    await page.waitForURL(/\/(fo|dashboard)/, { timeout: 15000 });

    check(page, { 'login: redirect ke dashboard': () => page.url().includes('/fo') });

    // ── 2. Buka halaman reservasi ─────────────────────────────────────────────
    const reservationPageStart = Date.now();
    await page.goto(`${BASE}/fo/reservations`, { waitUntil: 'networkidle' });
    pageLoadDuration.add(Date.now() - reservationPageStart);

    check(page, { 'halaman reservasi: terbuka': () => page.url().includes('/fo/reservations') });

    // ── 3. Klik tombol buat reservasi baru ────────────────────────────────────
    // RoomMasterb kemungkinan punya tombol "New Reservation" atau "Tambah"
    // Coba beberapa kemungkinan selector
    const newBtnSelectors = [
      'button:has-text("New Reservation")',
      'button:has-text("Tambah Reservasi")',
      'button:has-text("Add Reservation")',
      '[data-testid="new-reservation"]',
      'a[href*="new"]',
    ];

    let btnFound = false;
    for (const sel of newBtnSelectors) {
      try {
        await page.waitForSelector(sel, { timeout: 3000 });
        await page.locator(sel).click();
        btnFound = true;
        break;
      } catch {
        // coba selector berikutnya
      }
    }

    if (!btnFound) {
      // Fallback: navigasi langsung ke halaman booking baru
      await page.goto(`${BASE}/fo/reservations/new`, { waitUntil: 'networkidle' });
    }

    // ── 4. Isi form reservasi ─────────────────────────────────────────────────
    // Tunggu form muncul
    await page.waitForSelector('form, [role="dialog"]', { timeout: 10000 });

    // Isi tanggal check-in (format bisa berbeda tergantung komponen)
    const checkInSelectors = [
      'input[name="check_in_date"]',
      'input[name="checkIn"]',
      'input[placeholder*="Check In"]',
      'input[placeholder*="check-in"]',
    ];
    for (const sel of checkInSelectors) {
      try {
        await page.waitForSelector(sel, { timeout: 2000 });
        await page.locator(sel).fill(checkIn);
        break;
      } catch { /* next */ }
    }

    // Isi tanggal check-out
    const checkOutSelectors = [
      'input[name="check_out_date"]',
      'input[name="checkOut"]',
      'input[placeholder*="Check Out"]',
      'input[placeholder*="check-out"]',
    ];
    for (const sel of checkOutSelectors) {
      try {
        await page.waitForSelector(sel, { timeout: 2000 });
        await page.locator(sel).fill(checkOut);
        break;
      } catch { /* next */ }
    }

    // Pilih kamar (jika ada dropdown)
    try {
      await page.waitForSelector('select[name="room_id"], [data-testid="room-select"]', { timeout: 3000 });
      await page.locator('select[name="room_id"]').selectOption(roomId);
    } catch { /* kamar mungkin dipilih dengan cara lain */ }

    // ── 5. Submit form ────────────────────────────────────────────────────────
    const submitStart = Date.now();
    await waitAndClick(page, 'button[type="submit"]:not([disabled])');

    // Tunggu konfirmasi berhasil (toast, redirect, atau perubahan halaman)
    try {
      await page.waitForSelector(
        '[data-testid="success"], .mantine-Notification-root, [role="alert"]',
        { timeout: 10000 },
      );
      flowSuccess = true;
      bookingTotal.add(1);
    } catch {
      // Cek apakah halaman berubah (redirect setelah submit)
      const currentUrl = page.url();
      if (!currentUrl.includes('/new')) {
        flowSuccess = true;
        bookingTotal.add(1);
      }
    }
    pageLoadDuration.add(Date.now() - submitStart);

    check(page, { 'booking: berhasil dibuat': () => flowSuccess });

  } catch (err) {
    console.error(`[VU ${__VU}] Error: ${err.message}`);

    // Screenshot saat gagal untuk debugging
    try {
      await page.screenshot({ path: `k6/results/error_vu${__VU}_${Date.now()}.png` });
    } catch { /* ignore */ }

  } finally {
    bookingSuccessRate.add(flowSuccess);
    bookingFlowDuration.add(Date.now() - flowStart);
    await page.close();
  }

  sleep(2);
}

// ─── Summary ──────────────────────────────────────────────────────────────────
export function handleSummary(data) {
  const metrics = data.metrics;

  const p95Duration  = metrics.booking_flow_duration_ms?.values?.['p(95)'] ?? 0;
  const p95PageLoad  = metrics.page_load_duration_ms?.values?.['p(95)'] ?? 0;
  const successRate  = (metrics.booking_success_rate?.values?.rate ?? 0) * 100;
  const totalBookings = metrics.booking_total?.values?.count ?? 0;

  console.log('\n╔══════════════════════════════════════════╗');
  console.log('║   ROOMMASTER — HASIL BOOKING TEST        ║');
  console.log('╠══════════════════════════════════════════╣');
  console.log(`║ Booking sukses rate : ${successRate.toFixed(1).padStart(6)}%             ║`);
  console.log(`║ Total booking       : ${String(totalBookings).padStart(6)} kali           ║`);
  console.log(`║ Flow duration P95   : ${p95Duration.toFixed(0).padStart(6)} ms            ║`);
  console.log(`║ Page load P95       : ${p95PageLoad.toFixed(0).padStart(6)} ms            ║`);
  console.log('╚══════════════════════════════════════════╝\n');

  return {
    'k6/results/roommaster_summary.json': JSON.stringify(data, null, 2),
  };
}
