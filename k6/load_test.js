/**
 * LOAD TESTING — Hotel Booking API
 * Simulasi beban normal yang stabil (sustained load).
 *
 * Jalankan dari folder k6/:
 *   k6 run .\load_test.js
 *
 * Untuk RoomMasterb: ganti import flow, lalu:
 *   k6 run -e TARGET=roommaster .\load_test.js
 */

import { sleep } from "k6";
import { Rate, Trend, Counter } from "k6/metrics";
import { setup as hbSetup, runBookingFlow as hbFlow } from "./flow/hotelbooking.js";
import { setup as rmSetup, runBookingFlow as rmFlow } from "./flow/roommaster.js";

const TARGET = __ENV.TARGET || "hotelbooking";
const isRM   = TARGET === "roommaster";

export function setup() {
  return isRM ? rmSetup() : hbSetup();
}

const successRate = new Rate("booking_success_rate");
const flowDuration = new Trend("booking_flow_duration_ms", true);
const totalBookings = new Counter("booking_total");
const totalFailed = new Counter("booking_failed");

export const options = {
  summaryTrendStats: ["avg", "min", "med", "max", "p(90)", "p(95)", "p(99)"],
  scenarios: {
    load: {
      executor: "ramping-vus",
      startVUs: 0,
      stages: [
        { duration: "1m", target: 20 },
        { duration: "3m", target: 20 },
        { duration: "1m", target: 0 },
      ],
    },
  },
  thresholds: {
    http_req_duration: ["p(95)<3000"],
    http_req_failed: ["rate<0.05"],
    booking_success_rate: ["rate>0.90"],
    booking_flow_duration_ms: ["p(95)<8000"],
  },
};

export default function (data) {
  const result = isRM ? rmFlow(data.token) : hbFlow(data.token);

  successRate.add(result.success);
  flowDuration.add(result.durationMs);

  if (result.success) {
    totalBookings.add(1);
  } else {
    totalFailed.add(1);
  }

  sleep(1);
}

export function handleSummary(data) {
  const m = data.metrics;
  const success = ((m.booking_success_rate?.values?.rate ?? 0) * 100).toFixed(
    1,
  );
  const total = m.booking_total?.values?.count ?? 0;
  const failed = m.booking_failed?.values?.count ?? 0;
  const flowP95 = (m.booking_flow_duration_ms?.values?.["p(95)"] ?? 0).toFixed(
    0,
  );
  const httpMed = (m.http_req_duration?.values?.med ?? 0).toFixed(0);
  const httpP95 = (m.http_req_duration?.values?.["p(95)"] ?? 0).toFixed(0);
  const httpP99 = (m.http_req_duration?.values?.["p(99)"] ?? 0).toFixed(0);
  const reqTotal = m.http_reqs?.values?.count ?? 0;
  const reqRate = (m.http_reqs?.values?.rate ?? 0).toFixed(2);
  const errRate = ((m.http_req_failed?.values?.rate ?? 0) * 100).toFixed(2);

  console.log("\n╔══════════════════════════════════════════════════╗");
  console.log("║           LOAD TEST — HASIL PENGUJIAN            ║");
  console.log("╠══════════════════════════════════════════════════╣");
  console.log(`║  Target API      : ${TARGET.padEnd(29)}║`);
  console.log("╠══════════════════════════════════════════════════╣");
  console.log(
    `║  Total HTTP Req  : ${String(reqTotal).padStart(8)}                   ║`,
  );
  console.log(
    `║  Req/detik       : ${String(reqRate).padStart(8)}                   ║`,
  );
  console.log(
    `║  Error Rate      : ${String(errRate + "%").padStart(8)}                   ║`,
  );
  console.log("╠══════════════════════════════════════════════════╣");
  console.log(
    `║  HTTP P50 (med)  : ${String(httpMed + " ms").padStart(8)}                   ║`,
  );
  console.log(
    `║  HTTP P95        : ${String(httpP95 + " ms").padStart(8)}                   ║`,
  );
  console.log(
    `║  HTTP P99        : ${String(httpP99 + " ms").padStart(8)}                   ║`,
  );
  console.log("╠══════════════════════════════════════════════════╣");
  console.log(
    `║  Booking sukses  : ${String(success + "%").padStart(8)}                   ║`,
  );
  console.log(
    `║  Total booking   : ${String(total).padStart(8)}                   ║`,
  );
  console.log(
    `║  Total gagal     : ${String(failed).padStart(8)}                   ║`,
  );
  console.log(
    `║  Flow P95        : ${String(flowP95 + " ms").padStart(8)}                   ║`,
  );
  console.log("╚══════════════════════════════════════════════════╝\n");

  return {
    [`results/load_${TARGET}.json`]: JSON.stringify(data, null, 2),
  };
}
