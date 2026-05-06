/**
 * STRESS TESTING — Hotel Booking API
 * ─────────────────────────────────────────────────────────────────────────────
 * Naikkan beban secara bertahap melampaui kapasitas normal.
 * Mengukur breaking point sistem dan kemampuan recovery.
 *
 * Jalankan untuk hotelbooking (Go):
 *   k6 run k6/stress_test.js
 *
 * Jalankan untuk RoomMasterb:
 *   k6 run -e TARGET=roommaster k6/stress_test.js
 *
 * Simpan hasil:
 *   k6 run --out json=k6/results/stress_hotelbooking.json k6/stress_test.js
 */

import { sleep } from 'k6';
import { Rate, Trend, Counter } from 'k6/metrics';
import { runBookingFlow } from './flow/hotelbooking.js';
// import { runBookingFlow } from './flow/roommaster.js';  // uncomment untuk RoomMasterb

// ─── Custom Metrics ───────────────────────────────────────────────────────────
const successRate   = new Rate('booking_success_rate');
const flowDuration  = new Trend('booking_flow_duration_ms', true);
const totalBookings = new Counter('booking_total');
const totalFailed   = new Counter('booking_failed');

// ─── Skenario Stress Testing ──────────────────────────────────────────────────
// Pola: naik bertahap setiap menit, melebihi batas normal
// Amati: di VU mana error rate mulai naik, di mana sistem "patah"
export const options = {
  scenarios: {
    stress: {
      executor: 'ramping-vus',
      startVUs: 0,
      stages: [
        { duration: '1m', target: 10  },  // normal
        { duration: '1m', target: 20  },  // di atas normal
        { duration: '1m', target: 40  },  // mulai berat
        { duration: '1m', target: 60  },  // stress
        { duration: '1m', target: 80  },  // heavy stress
        { duration: '1m', target: 100 },  // extreme
        { duration: '1m', target: 0   },  // recovery
      ],
    },
  },
  // Threshold lebih longgar — kita INGIN melihat sistem patah
  thresholds: {
    http_req_duration:       ['p(95)<10000'],
    http_req_failed:         ['rate<0.20'],
    booking_success_rate:    ['rate>0.70'],
    booking_flow_duration_ms: ['p(95)<20000'],
  },
};

// ─── Test Utama ───────────────────────────────────────────────────────────────
export default function () {
  const result = runBookingFlow();

  successRate.add(result.success);
  flowDuration.add(result.durationMs);

  if (result.success) {
    totalBookings.add(1);
  } else {
    totalFailed.add(1);
  }

  sleep(1);
}

// ─── Summary ──────────────────────────────────────────────────────────────────
export function handleSummary(data) {
  const m = data.metrics;
  const success  = ((m.booking_success_rate?.values?.rate   ?? 0) * 100).toFixed(1);
  const total    = m.booking_total?.values?.count           ?? 0;
  const failed   = m.booking_failed?.values?.count          ?? 0;
  const flowP95  = (m.booking_flow_duration_ms?.values?.['p(95)'] ?? 0).toFixed(0);
  const httpP50  = (m.http_req_duration?.values?.['p(50)']        ?? 0).toFixed(0);
  const httpP95  = (m.http_req_duration?.values?.['p(95)']        ?? 0).toFixed(0);
  const httpP99  = (m.http_req_duration?.values?.['p(99)']        ?? 0).toFixed(0);
  const reqTotal = m.http_reqs?.values?.count               ?? 0;
  const reqRate  = (m.http_reqs?.values?.rate               ?? 0).toFixed(2);
  const errRate  = ((m.http_req_failed?.values?.rate        ?? 0) * 100).toFixed(2);

  console.log('\n╔══════════════════════════════════════════════════╗');
  console.log('║          STRESS TEST — HASIL PENGUJIAN           ║');
  console.log('╠══════════════════════════════════════════════════╣');
  console.log(`║  Target API      : hotelbooking (Go)             ║`);
  console.log('╠══════════════════════════════════════════════════╣');
  console.log(`║  Total HTTP Req  : ${String(reqTotal).padStart(8)}                   ║`);
  console.log(`║  Req/detik       : ${String(reqRate).padStart(8)}                   ║`);
  console.log(`║  Error Rate      : ${String(errRate + '%').padStart(8)}                   ║`);
  console.log('╠══════════════════════════════════════════════════╣');
  console.log(`║  HTTP P50        : ${String(httpP50 + ' ms').padStart(8)}                   ║`);
  console.log(`║  HTTP P95        : ${String(httpP95 + ' ms').padStart(8)}                   ║`);
  console.log(`║  HTTP P99        : ${String(httpP99 + ' ms').padStart(8)}                   ║`);
  console.log('╠══════════════════════════════════════════════════╣');
  console.log(`║  Booking sukses  : ${String(success + '%').padStart(8)}                   ║`);
  console.log(`║  Total booking   : ${String(total).padStart(8)}                   ║`);
  console.log(`║  Total gagal     : ${String(failed).padStart(8)}                   ║`);
  console.log(`║  Flow P95        : ${String(flowP95 + ' ms').padStart(8)}                   ║`);
  console.log('╚══════════════════════════════════════════════════╝\n');

  return {
    'k6/results/stress_hotelbooking.json': JSON.stringify(data, null, 2),
  };
}
