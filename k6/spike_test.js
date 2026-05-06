/**
 * SPIKE TESTING — Hotel Booking API
 * ─────────────────────────────────────────────────────────────────────────────
 * Simulasi lonjakan tiba-tiba jumlah pengguna.
 * Mengukur kemampuan sistem pulih dari traffic burst mendadak.
 *
 * Jalankan untuk hotelbooking (Go):
 *   k6 run k6/spike_test.js
 *
 * Jalankan untuk RoomMasterb:
 *   k6 run -e TARGET=roommaster k6/spike_test.js
 *
 * Simpan hasil:
 *   k6 run --out json=k6/results/spike_hotelbooking.json k6/spike_test.js
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

// ─── Skenario Spike Testing ───────────────────────────────────────────────────
// Pola: normal → spike tiba-tiba → kembali normal
// Mengukur: apakah sistem recover setelah spike, berapa error saat spike
export const options = {
  scenarios: {
    spike: {
      executor: 'ramping-vus',
      startVUs: 0,
      stages: [
        { duration: '30s', target: 5  },  // baseline normal
        { duration: '30s', target: 5  },  // tahan baseline
        { duration: '10s', target: 50 },  // SPIKE mendadak ke 50 VU
        { duration: '1m',  target: 50 },  // tahan spike 1 menit
        { duration: '10s', target: 5  },  // turun kembali ke normal
        { duration: '30s', target: 5  },  // recovery check
        { duration: '10s', target: 0  },  // selesai
      ],
    },
  },
  thresholds: {
    http_req_duration:       ['p(95)<5000'],   // lebih longgar karena spike
    http_req_failed:         ['rate<0.10'],    // toleransi error 10% saat spike
    booking_success_rate:    ['rate>0.80'],
    booking_flow_duration_ms: ['p(95)<12000'],
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
  console.log('║           SPIKE TEST — HASIL PENGUJIAN           ║');
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
    'k6/results/spike_hotelbooking.json': JSON.stringify(data, null, 2),
  };
}
