"""
analyze_ai_booking.py — Bandingkan AI booking (agentic) vs direct booking
(manual API calls). Untuk paper: menunjukkan trade-off antara UX natural
language dan latency/cost.

Prasyarat — sudah jalankan:
    k6 run k6/ai_booking_test.js                       (untuk AI booking)
    k6 run k6/endpoint_breakdown.js                    (untuk direct booking)

Cara pakai:
    python3 k6/analyze_ai_booking.py

Output:
    k6/results/ai_vs_direct_table.txt
    k6/results/chart_ai_vs_direct.png
"""
import json
import sys
from pathlib import Path

RESULT_DIR = Path(__file__).parent / "results"


def load_json(name):
    f = RESULT_DIR / name
    if not f.exists():
        return None
    try:
        return json.loads(f.read_text())
    except json.JSONDecodeError:
        return None


# ─── Load AI booking results ─────────────────────────────────────────────────
ai_data = load_json("ai_booking.json")
if not ai_data:
    print("ERROR: k6/results/ai_booking.json tidak ada.")
    print("Jalankan dulu: k6 run k6/ai_booking_test.js")
    sys.exit(1)

m = ai_data.get("metrics", {})


def val(key, stat):
    return m.get(key, {}).get("values", {}).get(stat, 0)


ai_stats = {
    "lat_avg": val("ai_booking_duration_ms", "avg"),
    "lat_med": val("ai_booking_duration_ms", "med"),
    "lat_p95": val("ai_booking_duration_ms", "p(95)"),
    "lat_p99": val("ai_booking_duration_ms", "p(99)"),
    "lat_max": val("ai_booking_duration_ms", "max"),
    "iter_avg": val("ai_openai_iterations", "avg"),
    "iter_max": val("ai_openai_iterations", "max"),
    "tools_avg": val("ai_tool_calls_count", "avg"),
    "tools_max": val("ai_tool_calls_count", "max"),
    "success": val("ai_booking_success", "rate") * 100,
    "total_success": val("ai_total_success", "count"),
    "total_failed": val("ai_total_failed", "count"),
}

# ─── Load direct booking (endpoint_breakdown) ────────────────────────────────
direct_data = load_json("endpoint_breakdown_hotelbooking.json")
direct_stats = None
if direct_data:
    dm = direct_data.get("metrics", {})

    def dval(key, stat):
        return dm.get(key, {}).get("values", {}).get(stat, 0)

    # Total flow direct = check_availability + create + pay (tanpa cancel,
    # karena AI booking juga sampai pay)
    flow_avg = dval("ep_check_availability_ms", "avg") + dval("ep_create_reservation_ms", "avg") + dval("ep_pay_reservation_ms", "avg")
    flow_med = dval("ep_check_availability_ms", "med") + dval("ep_create_reservation_ms", "med") + dval("ep_pay_reservation_ms", "med")
    flow_p95 = dval("ep_check_availability_ms", "p(95)") + dval("ep_create_reservation_ms", "p(95)") + dval("ep_pay_reservation_ms", "p(95)")
    flow_p99 = dval("ep_check_availability_ms", "p(99)") + dval("ep_create_reservation_ms", "p(99)") + dval("ep_pay_reservation_ms", "p(99)")
    flow_max = dval("ep_check_availability_ms", "max") + dval("ep_create_reservation_ms", "max") + dval("ep_pay_reservation_ms", "max")

    direct_stats = {
        "lat_avg": flow_avg,
        "lat_med": flow_med,
        "lat_p95": flow_p95,
        "lat_p99": flow_p99,
        "lat_max": flow_max,
    }


# ─── Cost calculation (gpt-4o-mini, asumsi 1500 input + 200 output token) ────
PRICE_INPUT_PER_1M = 0.15
PRICE_OUTPUT_PER_1M = 0.60
TOKENS_IN_PER_CALL = 1500
TOKENS_OUT_PER_CALL = 200

cost_per_call = (TOKENS_IN_PER_CALL / 1e6) * PRICE_INPUT_PER_1M + (TOKENS_OUT_PER_CALL / 1e6) * PRICE_OUTPUT_PER_1M
cost_per_booking = cost_per_call * ai_stats["iter_avg"]
cost_1000_bookings = cost_per_booking * 1000


# ─── Tabel teks ──────────────────────────────────────────────────────────────
lines = []
sep = "=" * 80
lines.append(sep)
lines.append("  AI BOOKING vs DIRECT BOOKING — Trade-off Analysis")
lines.append(sep)
lines.append("")
lines.append("  [ LATENCY (end-to-end per booking) ]")
lines.append("  " + "-" * 78)
lines.append(f"  {'Metric':<18}{'AI Booking':>18}{'Direct Booking':>20}{'Ratio':>12}")
lines.append("  " + "-" * 78)
if direct_stats:
    for label, key in [("Avg", "lat_avg"), ("Median", "lat_med"), ("P95", "lat_p95"), ("P99", "lat_p99"), ("Max", "lat_max")]:
        ai_v = ai_stats[key]
        dr_v = direct_stats[key]
        ratio = (ai_v / dr_v) if dr_v > 0 else 0
        lines.append(f"  {label:<18}{ai_v:>14.0f} ms{dr_v:>16.0f} ms{ratio:>10.1f}×")
else:
    for label, key in [("Avg", "lat_avg"), ("Median", "lat_med"), ("P95", "lat_p95"), ("P99", "lat_p99"), ("Max", "lat_max")]:
        lines.append(f"  {label:<18}{ai_stats[key]:>14.0f} ms{'N/A':>20}{'-':>12}")
    lines.append("  (Direct booking data tidak ada — run k6 run k6/endpoint_breakdown.js)")

lines.append("")
lines.append("  [ AI OVERHEAD ]")
lines.append("  " + "-" * 78)
lines.append(f"  OpenAI iterations (avg)      : {ai_stats['iter_avg']:.2f}  (max: {ai_stats['iter_max']:.0f})")
lines.append(f"  Tool calls per booking (avg) : {ai_stats['tools_avg']:.2f}  (max: {ai_stats['tools_max']:.0f})")
lines.append(f"  Success rate                 : {ai_stats['success']:.1f}%")
lines.append(f"  Total: {int(ai_stats['total_success'])} sukses, {int(ai_stats['total_failed'])} gagal")
lines.append("")
lines.append("  [ COST ESTIMATE (gpt-4o-mini) ]")
lines.append("  " + "-" * 78)
lines.append(f"  Asumsi: {TOKENS_IN_PER_CALL} input + {TOKENS_OUT_PER_CALL} output token/call")
lines.append(f"  Pricing: ${PRICE_INPUT_PER_1M}/1M input, ${PRICE_OUTPUT_PER_1M}/1M output")
lines.append(f"  Cost per OpenAI call         : ~${cost_per_call:.6f}")
lines.append(f"  Cost per AI booking          : ~${cost_per_booking:.6f}")
lines.append(f"  Cost per 1.000 booking       : ~${cost_1000_bookings:.3f}")
lines.append(f"  Cost per 100.000 booking     : ~${cost_1000_bookings * 100:.2f}")
lines.append("")
lines.append(sep)

table = "\n".join(lines)
print(table)
(RESULT_DIR / "ai_vs_direct_table.txt").write_text(table + "\n")
print(f"\nTabel disimpan: {RESULT_DIR / 'ai_vs_direct_table.txt'}")


# ─── Chart ───────────────────────────────────────────────────────────────────
try:
    import matplotlib
    matplotlib.use("Agg")
    import matplotlib.pyplot as plt
except ImportError:
    print("matplotlib tidak terinstall — chart dilewati.")
    sys.exit(0)

if not direct_stats:
    print("Direct booking data tidak ada — chart komparasi dilewati.")
    sys.exit(0)

# Gambar dengan 2 panel: latency comparison + cost projection
fig, axes = plt.subplots(1, 2, figsize=(14, 5.5))

# ─── Panel kiri: Latency comparison ──────────────────────────────────────────
ax = axes[0]
metrics_lbl = ["Avg", "Median", "P95", "P99", "Max"]
ai_vals = [ai_stats["lat_avg"], ai_stats["lat_med"], ai_stats["lat_p95"], ai_stats["lat_p99"], ai_stats["lat_max"]]
dr_vals = [direct_stats["lat_avg"], direct_stats["lat_med"], direct_stats["lat_p95"], direct_stats["lat_p99"], direct_stats["lat_max"]]

import numpy as np
x = np.arange(len(metrics_lbl))
width = 0.4

bars1 = ax.bar(x - width/2, ai_vals, width, label="AI Booking", color="#9C27B0", alpha=0.9)
bars2 = ax.bar(x + width/2, dr_vals, width, label="Direct Booking", color="#0D47A1", alpha=0.9)

for bar, v in zip(bars1, ai_vals):
    ax.text(bar.get_x() + bar.get_width()/2, bar.get_height(), f"{v:.0f}",
            ha="center", va="bottom", fontsize=8)
for bar, v in zip(bars2, dr_vals):
    ax.text(bar.get_x() + bar.get_width()/2, bar.get_height(), f"{v:.0f}",
            ha="center", va="bottom", fontsize=8)

ax.set_ylim(0, max(ai_vals + dr_vals) * 1.18)
ax.set_title("Latency: AI Booking vs Direct Booking", fontsize=13, fontweight="bold", pad=8)
ax.set_ylabel("Latency (ms)", fontsize=11)
ax.set_xlabel("Metric", fontsize=11)
ax.set_xticks(x)
ax.set_xticklabels(metrics_lbl, fontsize=10)
ax.grid(axis="y", alpha=0.3, linestyle="--")
ax.legend(fontsize=10)
ax.annotate("↓ better", xy=(0.98, 0.97), xycoords="axes fraction",
            ha="right", va="top", fontsize=9, color="gray", style="italic")

# ─── Panel kanan: Cost projection ────────────────────────────────────────────
ax = axes[1]
scales = ["100", "1K", "10K", "100K", "1M"]
multipliers = [100, 1_000, 10_000, 100_000, 1_000_000]
costs = [cost_per_booking * n for n in multipliers]

bars = ax.bar(scales, costs, color="#FF9800", alpha=0.9)
for bar, v in zip(bars, costs):
    label = f"${v:.4f}" if v < 1 else f"${v:.2f}"
    ax.text(bar.get_x() + bar.get_width()/2, bar.get_height(), label,
            ha="center", va="bottom", fontsize=10)

ax.set_ylim(0, max(costs) * 1.15)
ax.set_title(f"Estimasi Biaya OpenAI per Volume Booking\n(gpt-4o-mini, {ai_stats['iter_avg']:.1f} iter avg)",
             fontsize=12, fontweight="bold", pad=8)
ax.set_ylabel("Cost (USD)", fontsize=11)
ax.set_xlabel("Jumlah Booking", fontsize=11)
ax.grid(axis="y", alpha=0.3, linestyle="--")

plt.tight_layout()
out = RESULT_DIR / "chart_ai_vs_direct.png"
plt.savefig(out, dpi=150, bbox_inches="tight")
plt.close()
print(f"Chart disimpan: {out}")

print(f"\nSelesai! Semua file ada di: {RESULT_DIR}")
