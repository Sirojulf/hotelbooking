"""
analyze.py — Bandingkan hasil k6 antara hotelbooking (Go) dan RoomMasterb
Menghasilkan tabel perbandingan dan chart bar seperti di referensi PDF.

Cara pakai:
    python k6/analyze.py

Dependensi:
    pip install matplotlib pandas

Output:
    k6/results/comparison_table.txt
    k6/results/chart_success_requests.png
    k6/results/chart_response_time.png
    k6/results/chart_error_rate.png
"""

import json
import os
import sys
from pathlib import Path

RESULT_DIR = Path(__file__).parent / "results"

# ─── Coba import matplotlib ───────────────────────────────────────────────────
try:
    import matplotlib
    matplotlib.use("Agg")  # non-interactive backend
    import matplotlib.pyplot as plt
    import matplotlib.patches as mpatches
    HAS_MATPLOTLIB = True
except ImportError:
    HAS_MATPLOTLIB = False
    print("INFO: matplotlib tidak terinstall. Hanya tabel teks yang akan dibuat.")
    print("      Install dengan: pip install matplotlib")

# ─── Load JSON hasil k6 ───────────────────────────────────────────────────────
def load_k6_result(filepath: Path) -> dict:
    """Baca file JSON output k6 dan ekstrak metrik utama."""
    if not filepath.exists():
        return None
    with open(filepath) as f:
        data = json.load(f)

    metrics = data.get("metrics", {})

    def val(key, stat):
        return metrics.get(key, {}).get("values", {}).get(stat, 0)

    return {
        "total_requests":    val("http_reqs", "count"),
        "req_per_sec":       round(val("http_reqs", "rate"), 2),
        "error_rate_pct":    round(val("http_req_failed", "rate") * 100, 2),
        "success_rate_pct":  round(val("booking_success_rate", "rate") * 100, 1),
        "total_bookings":    val("booking_total", "count"),
        "total_failed":      val("booking_failed", "count"),
        "http_p50_ms":       round(val("http_req_duration", "p(50)"), 0),
        "http_p95_ms":       round(val("http_req_duration", "p(95)"), 0),
        "http_p99_ms":       round(val("http_req_duration", "p(99)"), 0),
        "flow_p95_ms":       round(val("booking_flow_duration_ms", "p(95)"), 0),
    }

def load_monitor_summary(test_name: str, target: str) -> dict:
    """Baca file ringkasan CPU/Mem dari monitor.ps1."""
    filepath = RESULT_DIR / f"{target}_{test_name}_cpu_mem_summary.txt"
    result = {"cpu_avg": "N/A", "cpu_max": "N/A", "mem_avg": "N/A", "mem_max": "N/A"}
    if not filepath.exists():
        return result
    for line in filepath.read_text().splitlines():
        if "CPU Avg" in line:
            result["cpu_avg"] = line.split(":")[-1].strip()
        elif "CPU Max" in line:
            result["cpu_max"] = line.split(":")[-1].strip()
        elif "Mem Avg" in line:
            result["mem_avg"] = line.split(":")[-1].strip()
        elif "Mem Max" in line:
            result["mem_max"] = line.split(":")[-1].strip()
    return result

# ─── Kumpulkan semua data ─────────────────────────────────────────────────────
TARGETS    = ["hotelbooking", "roommaster"]
TEST_TYPES = ["load", "spike", "stress"]
TARGET_LABELS = {
    "hotelbooking": "Go (hotelbooking)",
    "roommaster":   "Node.js (RoomMasterb)",
}

results = {}
for target in TARGETS:
    results[target] = {}
    for test in TEST_TYPES:
        filepath = RESULT_DIR / f"{test}_{target}.json"
        r = load_k6_result(filepath)
        if r:
            monitor = load_monitor_summary(test, target)
            r.update(monitor)
            results[target][test] = r
        else:
            results[target][test] = None

# ─── Print tabel teks ─────────────────────────────────────────────────────────
def print_table():
    sep  = "=" * 90
    sep2 = "-" * 90

    print(f"\n{sep}")
    print(f"  TABEL PERBANDINGAN KINERJA: hotelbooking (Go) vs RoomMasterb (Node.js)")
    print(f"{sep}")

    for test in TEST_TYPES:
        print(f"\n  [ {test.upper()} TESTING ]")
        print(f"  {sep2}")
        header = f"  {'Metrik':<30} {'Go (hotelbooking)':>22} {'Node.js (RoomMasterb)':>22}"
        print(header)
        print(f"  {sep2}")

        metrics_list = [
            ("Total HTTP Requests",   "total_requests",   ""),
            ("Req / detik",           "req_per_sec",      "rps"),
            ("Error Rate",            "error_rate_pct",   "%"),
            ("Booking Success Rate",  "success_rate_pct", "%"),
            ("Total Booking Sukses",  "total_bookings",   ""),
            ("Total Booking Gagal",   "total_failed",     ""),
            ("HTTP Response P50",     "http_p50_ms",      "ms"),
            ("HTTP Response P95",     "http_p95_ms",      "ms"),
            ("HTTP Response P99",     "http_p99_ms",      "ms"),
            ("Flow Duration P95",     "flow_p95_ms",      "ms"),
            ("CPU Rata-rata",         "cpu_avg",          ""),
            ("CPU Maksimum",          "cpu_max",          ""),
            ("Memory Rata-rata",      "mem_avg",          ""),
            ("Memory Maksimum",       "mem_max",          ""),
        ]

        for label, key, unit in metrics_list:
            hb_val  = results.get("hotelbooking", {}).get(test)
            rm_val  = results.get("roommaster", {}).get(test)

            hb_str  = f"{hb_val[key]}{unit}" if hb_val and hb_val.get(key) is not None else "N/A"
            rm_str  = f"{rm_val[key]}{unit}" if rm_val and rm_val.get(key) is not None else "N/A"

            print(f"  {label:<30} {hb_str:>22} {rm_str:>22}")

        print(f"  {sep2}")

    print(f"\n{sep}\n")

print_table()

# Simpan tabel ke file
table_file = RESULT_DIR / "comparison_table.txt"
RESULT_DIR.mkdir(exist_ok=True)
import io, contextlib
buf = io.StringIO()
with contextlib.redirect_stdout(buf):
    print_table()
table_file.write_text(buf.getvalue())
print(f"Tabel disimpan ke: {table_file}")

# ─── Buat chart ───────────────────────────────────────────────────────────────
if not HAS_MATPLOTLIB:
    print("Selesai (tanpa chart).")
    sys.exit(0)

COLORS = {"hotelbooking": "#2196F3", "roommaster": "#FF5722"}
X      = list(range(len(TEST_TYPES)))
BAR_W  = 0.35

def make_chart(metric_key, ylabel, title, filename, unit="", higher_is_better=True):
    fig, ax = plt.subplots(figsize=(9, 5))

    for i, target in enumerate(TARGETS):
        vals = []
        for test in TEST_TYPES:
            r = results[target].get(test)
            try:
                v = float(str(r[metric_key]).replace("%", "").replace("ms", "").strip()) if r else 0
            except (TypeError, ValueError):
                v = 0
            vals.append(v)

        offset = (i - 0.5) * BAR_W
        bars = ax.bar([x + offset for x in X], vals, BAR_W,
                      label=TARGET_LABELS[target], color=COLORS[target], alpha=0.85)

        for bar, v in zip(bars, vals):
            ax.text(bar.get_x() + bar.get_width() / 2, bar.get_height() + 0.5,
                    f"{v:.0f}{unit}", ha="center", va="bottom", fontsize=8)

    ax.set_xticks(X)
    ax.set_xticklabels([t.capitalize() for t in TEST_TYPES])
    ax.set_xlabel("Jenis Test")
    ax.set_ylabel(ylabel)
    ax.set_title(title)
    ax.legend()
    ax.grid(axis="y", alpha=0.3)

    note = "Lebih tinggi = lebih baik" if higher_is_better else "Lebih rendah = lebih baik"
    ax.annotate(note, xy=(0.99, 0.01), xycoords="axes fraction",
                ha="right", va="bottom", fontsize=7, color="gray")

    plt.tight_layout()
    outpath = RESULT_DIR / filename
    plt.savefig(outpath, dpi=150)
    plt.close()
    print(f"Chart disimpan: {outpath}")

make_chart("total_requests",   "Total HTTP Requests",    "Success Requests Total",         "chart_success_requests.png",  higher_is_better=True)
make_chart("req_per_sec",      "Requests / detik",       "Throughput (Req/s)",             "chart_throughput.png",        higher_is_better=True)
make_chart("error_rate_pct",   "Error Rate (%)",         "Error Rate",                     "chart_error_rate.png",  "%",  higher_is_better=False)
make_chart("http_p95_ms",      "Response Time P95 (ms)", "HTTP Response Time P95",         "chart_response_p95.png", "ms", higher_is_better=False)
make_chart("success_rate_pct", "Booking Success (%)",    "Booking Success Rate",           "chart_booking_success.png","%", higher_is_better=True)

print(f"\nSelesai! Semua file ada di: {RESULT_DIR}")
