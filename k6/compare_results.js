/**
 * Baca dua file hasil JSON dari k6 dan tampilkan tabel perbandingan.
 *
 * Jalankan setelah kedua test selesai:
 *   node k6/compare_results.js
 */

const fs = require('fs');
const path = require('path');

function loadSummary(filename) {
  const filepath = path.join(__dirname, 'results', filename);
  if (!fs.existsSync(filepath)) {
    console.error(`File tidak ditemukan: ${filepath}`);
    console.error('Pastikan kamu sudah menjalankan kedua k6 test terlebih dahulu.');
    process.exit(1);
  }
  return JSON.parse(fs.readFileSync(filepath, 'utf8'));
}

function getMetric(data, name, stat = 'p(95)') {
  const m = data.metrics[name];
  if (!m) return 'N/A';
  const val = m.values[stat] ?? m.values['rate'] ?? m.values['count'] ?? 0;
  return typeof val === 'number' ? val.toFixed(2) : val;
}

function bar(val, max, width = 30) {
  const filled = Math.round((val / max) * width);
  return '█'.repeat(filled) + '░'.repeat(width - filled);
}

const hb = loadSummary('hotelbooking_summary.json');
const rm = loadSummary('roommaster_summary.json');

// ─── Kumpulkan data ───────────────────────────────────────────────────────────
const data = [
  {
    metric:   'Booking Sukses Rate',
    unit:     '%',
    hb:       parseFloat(getMetric(hb, 'booking_success_rate', 'rate')) * 100,
    rm:       parseFloat(getMetric(rm, 'booking_success_rate', 'rate')) * 100,
    higherIsBetter: true,
  },
  {
    metric:   'Total Booking',
    unit:     'kali',
    hb:       parseFloat(getMetric(hb, 'booking_total', 'count')),
    rm:       parseFloat(getMetric(rm, 'booking_total', 'count')),
    higherIsBetter: true,
  },
  {
    metric:   'Flow Duration P50',
    unit:     'ms',
    hb:       parseFloat(getMetric(hb, 'booking_flow_duration_ms', 'p(50)')),
    rm:       parseFloat(getMetric(rm, 'booking_flow_duration_ms', 'p(50)')),
    higherIsBetter: false,
  },
  {
    metric:   'Flow Duration P95',
    unit:     'ms',
    hb:       parseFloat(getMetric(hb, 'booking_flow_duration_ms', 'p(95)')),
    rm:       parseFloat(getMetric(rm, 'booking_flow_duration_ms', 'p(95)')),
    higherIsBetter: false,
  },
  {
    metric:   'HTTP Req Duration P95',
    unit:     'ms',
    hb:       parseFloat(getMetric(hb, 'http_req_duration', 'p(95)')),
    rm:       parseFloat(getMetric(rm, 'http_req_duration', 'p(95)')),
    higherIsBetter: false,
  },
  {
    metric:   'HTTP Req Failed Rate',
    unit:     '%',
    hb:       parseFloat(getMetric(hb, 'http_req_failed', 'rate')) * 100,
    rm:       parseFloat(getMetric(rm, 'http_req_failed', 'rate')) * 100,
    higherIsBetter: false,
  },
];

// ─── Print tabel ──────────────────────────────────────────────────────────────
console.log('\n');
console.log('═'.repeat(72));
console.log('  KOMPARASI PERFORMA BOOKING FLOW');
console.log('  hotelbooking (Go REST API)  vs  RoomMasterb (Next.js SSR)');
console.log('═'.repeat(72));
console.log(
  '  Metrik'.padEnd(28) +
  'hotelbooking'.padStart(14) +
  'RoomMasterb'.padStart(14) +
  'Winner'.padStart(12),
);
console.log('─'.repeat(72));

for (const row of data) {
  const hbNum = isNaN(row.hb) ? 0 : row.hb;
  const rmNum = isNaN(row.rm) ? 0 : row.rm;

  let winner = '─';
  if (hbNum !== rmNum) {
    if (row.higherIsBetter) {
      winner = hbNum > rmNum ? '✅ hotelbooking' : '✅ RoomMasterb';
    } else {
      winner = hbNum < rmNum ? '✅ hotelbooking' : '✅ RoomMasterb';
    }
  }

  const hbStr = `${hbNum.toFixed(1)} ${row.unit}`;
  const rmStr = `${rmNum.toFixed(1)} ${row.unit}`;

  console.log(
    `  ${row.metric}`.padEnd(28) +
    hbStr.padStart(14) +
    rmStr.padStart(14) +
    winner.padStart(14),
  );
}

console.log('═'.repeat(72));
console.log('\n  Catatan:');
console.log('  • hotelbooking diukur sebagai pure REST API (HTTP load test)');
console.log('  • RoomMasterb diukur sebagai full UI flow (browser test)');
console.log('  • Perbedaan arsitektur membuat keduanya tidak apple-to-apple,');
console.log('    tapi menggambarkan real-world user experience masing-masing.\n');
