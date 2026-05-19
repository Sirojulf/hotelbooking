/**
 * AI BOOKING BENCHMARK — pengukuran agentic booking via OpenAI Function Calling.
 *
 * Endpoint: POST /api/v1/guests/ai/book
 *
 * Yang diukur:
 *   - End-to-end latency (1 request = beberapa OpenAI calls + Supabase RPC)
 *   - Jumlah iterasi OpenAI per booking (proxy biaya token)
 *   - Jumlah tool calls per booking
 *   - Booking success rate (apakah reservation_id dihasilkan)
 *   - Estimasi biaya per booking
 *
 * Catatan:
 *   - Low concurrency (2 VU) karena tiap booking lambat & makan API quota
 *   - Durasi pendek (3 menit) → ~30-60 booking attempts → cukup untuk statistik
 *   - Tanggal unik per (VU, iter) untuk hindari konflik
 *   - Auto-cancel setelah sukses (cleanup)
 *
 * Jalankan:
 *   k6 run k6/ai_booking_test.js
 */
import http from 'k6/http';
import { check, sleep } from 'k6';
import { Trend, Rate, Counter } from 'k6/metrics';
import { HOTELBOOKING, getCheckInOut } from './config.js';

// ─── Custom metrics ──────────────────────────────────────────────────────────
const aiLatency       = new Trend('ai_booking_duration_ms', true);
const aiIterations    = new Trend('ai_openai_iterations');
const aiToolCalls     = new Trend('ai_tool_calls_count');
const aiBookingRate   = new Rate('ai_booking_success');
const aiSuccessTotal  = new Counter('ai_total_success');
const aiFailedTotal   = new Counter('ai_total_failed');

export const options = {
  summaryTrendStats: ['avg', 'min', 'med', 'p(90)', 'p(95)', 'p(99)', 'max'],
  scenarios: {
    ai_booking: {
      executor: 'constant-vus',
      vus: 2,             // low — AI lambat + ada rate limit OpenAI
      duration: '3m',
    },
  },
  thresholds: {
    http_req_duration: ['p(95)<30000'],   // 30 detik budget per request
    ai_booking_success: ['rate>0.70'],    // minimal 70% sukses
  },
};

// ─── Setup: login sekali ─────────────────────────────────────────────────────
export function setup() {
  const res = http.post(
    `${HOTELBOOKING.baseURL}/auth/guest/login`,
    JSON.stringify({ login: HOTELBOOKING.email, password: HOTELBOOKING.password }),
    { headers: { 'Content-Type': 'application/json' } },
  );
  if (res.status !== 200) {
    console.error(`[setup] Login gagal: ${res.status} - ${res.body}`);
    return { token: null };
  }
  console.log('[setup] Login berhasil, token siap');
  return { token: res.json('access_token') };
}

// ─── Iterasi utama: 1 AI booking + auto-cancel ───────────────────────────────
export default function (data) {
  if (!data.token) return;

  const vuId = __VU - 1;
  const iter = __ITER;
  const { checkIn, checkOut } = getCheckInOut(vuId, iter);

  // Message ke AI — sengaja agak alami, sebutkan tanggal jelas supaya AI tahu
  // dia harus pakai tanggal itu (bukan tanggal hari ini).
  const message =
    `Tolong booking-kan saya kamar untuk 2 malam dari tanggal ${checkIn} ` +
    `sampai ${checkOut}. Cari hotel apa saja yang tersedia, pilih kamar ` +
    `yang paling murah, lalu langsung selesaikan: buat reservation dan ` +
    `lakukan pembayaran. Tidak perlu konfirmasi lagi ke saya.`;

  const headers = {
    'Content-Type': 'application/json',
    Authorization: `Bearer ${data.token}`,
  };

  // ─── Panggil AI booking endpoint ──────────────────────────────────────────
  const start = Date.now();
  const res = http.post(
    `${HOTELBOOKING.baseURL}/guests/ai/book`,
    JSON.stringify({ message }),
    { headers, timeout: '60s' },
  );
  const duration = Date.now() - start;
  aiLatency.add(duration);

  const httpOK = check(res, {
    'ai/book status 200': (r) => r.status === 200,
  });

  let reservationID = null;
  let iterations = 0;
  let toolCallsCount = 0;

  if (httpOK) {
    try {
      const body = res.json();
      iterations = body.iterations || 0;
      toolCallsCount = (body.tool_calls && body.tool_calls.length) || 0;
      reservationID = body.reservation_id || null;
    } catch (e) {
      console.error(`VU${__VU} iter${iter}: gagal parse response`);
    }
  }

  aiIterations.add(iterations);
  aiToolCalls.add(toolCallsCount);

  const bookingOK = !!reservationID;
  aiBookingRate.add(bookingOK);
  if (bookingOK) {
    aiSuccessTotal.add(1);

    // ─── Cleanup: cancel reservation supaya tidak menumpuk ──────────────────
    http.post(
      `${HOTELBOOKING.baseURL}/guests/reservations/${reservationID}/cancel`,
      null,
      { headers, timeout: '15s', tags: { phase: 'cleanup' } },
    );
  } else {
    aiFailedTotal.add(1);
  }

  sleep(1);
}

// ─── Summary akhir test ──────────────────────────────────────────────────────
export function handleSummary(data) {
  const m = data.metrics;

  const latAvg  = m.ai_booking_duration_ms?.values?.avg     ?? 0;
  const latMed  = m.ai_booking_duration_ms?.values?.med     ?? 0;
  const latP95  = m.ai_booking_duration_ms?.values?.['p(95)'] ?? 0;
  const latP99  = m.ai_booking_duration_ms?.values?.['p(99)'] ?? 0;
  const latMax  = m.ai_booking_duration_ms?.values?.max     ?? 0;

  const itAvg   = m.ai_openai_iterations?.values?.avg ?? 0;
  const itMax   = m.ai_openai_iterations?.values?.max ?? 0;
  const tcAvg   = m.ai_tool_calls_count?.values?.avg ?? 0;
  const tcMax   = m.ai_tool_calls_count?.values?.max ?? 0;

  const success = (m.ai_booking_success?.values?.rate ?? 0) * 100;
  const total   = m.ai_total_success?.values?.count ?? 0;
  const failed  = m.ai_total_failed?.values?.count ?? 0;

  // ─── Cost estimate (gpt-4o-mini) ─────────────────────────────────────────
  // Asumsi rata-rata 1500 input + 200 output token per OpenAI call
  // Pricing per Mei 2025: $0.15/1M input, $0.60/1M output
  const inputCost  = (1500 / 1_000_000) * 0.15;
  const outputCost = (200  / 1_000_000) * 0.60;
  const costPerCall = inputCost + outputCost;
  const costPerBooking = costPerCall * itAvg;

  const lines = [
    '',
    '╔══════════════════════════════════════════════════════════╗',
    '║          AI BOOKING — HASIL BENCHMARK                    ║',
    '╠══════════════════════════════════════════════════════════╣',
    `║  Latency avg          : ${pad(latAvg.toFixed(0), 7)} ms              ║`,
    `║  Latency med  (P50)   : ${pad(latMed.toFixed(0), 7)} ms              ║`,
    `║  Latency P95          : ${pad(latP95.toFixed(0), 7)} ms              ║`,
    `║  Latency P99          : ${pad(latP99.toFixed(0), 7)} ms              ║`,
    `║  Latency max          : ${pad(latMax.toFixed(0), 7)} ms              ║`,
    '╠══════════════════════════════════════════════════════════╣',
    `║  OpenAI iter avg      : ${pad(itAvg.toFixed(2), 7)}                 ║`,
    `║  OpenAI iter max      : ${pad(String(itMax), 7)}                 ║`,
    `║  Tool calls avg       : ${pad(tcAvg.toFixed(2), 7)}                 ║`,
    `║  Tool calls max       : ${pad(String(tcMax), 7)}                 ║`,
    '╠══════════════════════════════════════════════════════════╣',
    `║  Booking sukses rate  : ${pad(success.toFixed(1) + '%', 7)}              ║`,
    `║  Total booking sukses : ${pad(String(total), 7)}                 ║`,
    `║  Total booking gagal  : ${pad(String(failed), 7)}                 ║`,
    '╠══════════════════════════════════════════════════════════╣',
    `║  Estimasi biaya/call  : ~$${costPerCall.toFixed(6)}             ║`,
    `║  Estimasi biaya/booking: ~$${costPerBooking.toFixed(6)}            ║`,
    `║  Estimasi 1000 booking : ~$${(costPerBooking * 1000).toFixed(3)}              ║`,
    '╚══════════════════════════════════════════════════════════╝',
    '',
  ];
  for (const l of lines) console.log(l);

  return {
    'k6/results/ai_booking.json': JSON.stringify(data, null, 2),
  };
}

function pad(s, n) {
  return String(s).padStart(n);
}
