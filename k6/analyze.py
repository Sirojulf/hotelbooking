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
    import matplotlib.patches as mpatches
    import matplotlib.pyplot as plt

    HAS_MATPLOTLIB = True
except ImportError:
    HAS_MATPLOTLIB = False
    print("INFO: matplotlib tidak terinstall. Hanya tabel teks yang akan dibuat.")
    print("      Install dengan: pip install matplotlib")


# ─── Load JSON hasil k6 ───────────────────────────────────────────────────────
def load_k6_result(filepath: Path) -> dict:
    """Baca file JSON output k6 dan ekstrak metrik utama.
    Mendukung dua format:
      - Summary JSON  : dari handleSummary() di script k6 (format yang benar)
      - NDJSON        : dari flag --out json= (satu objek per baris, format lama)
    """
    if not filepath.exists():
        return None

    raw = filepath.read_text(encoding="utf-8").strip()
    if not raw:
        return None

    # Coba parse sebagai single JSON dulu (format summary dari handleSummary)
    try:
        data = json.loads(raw)
        if "metrics" not in data:
            print(f"WARN: {filepath.name} tidak punya key 'metrics', dilewati.")
            return None
        metrics = data["metrics"]
    except json.JSONDecodeError:
        # Fallback: NDJSON (--out json= format) — agregasi manual dari tiap baris
        print(
            f"INFO: {filepath.name} terdeteksi format NDJSON (dari --out json=), parsing..."
        )
        metrics = _parse_ndjson(raw, filepath)
        if metrics is None:
            return None

    def val(key, stat):
        return metrics.get(key, {}).get("values", {}).get(stat, 0)

    return {
        "total_requests": val("http_reqs", "count"),
        "req_per_sec": round(val("http_reqs", "rate"), 2),
        "error_rate_pct": round(val("http_req_failed", "rate") * 100, 2),
        "success_rate_pct": round(val("booking_success_rate", "rate") * 100, 1),
        "total_bookings": val("booking_total", "count"),
        "total_failed": val("booking_failed", "count"),
        "http_p50_ms": round(val("http_req_duration", "med"), 0),
        "http_p95_ms": round(val("http_req_duration", "p(95)"), 0),
        "http_p99_ms": round(val("http_req_duration", "p(99)"), 0),
        "flow_p95_ms": round(val("booking_flow_duration_ms", "p(95)"), 0),
    }


def _parse_ndjson(raw: str, filepath: Path) -> dict | None:
    """Parse format NDJSON dari k6 --out json= dan bentuk struktur metrics sederhana."""
    import math
    from collections import defaultdict

    counts = defaultdict(float)  # metric_name -> total count
    rates = defaultdict(list)  # metric_name -> list of rate values
    durations = defaultdict(list)  # metric_name -> list of duration values (ms)
    failed_count = 0
    total_reqs = 0

    for line in raw.splitlines():
        line = line.strip()
        if not line:
            continue
        try:
            obj = json.loads(line)
        except json.JSONDecodeError:
            continue

        if obj.get("type") != "Point":
            continue

        metric = obj.get("metric", "")
        val = obj.get("data", {}).get("value", 0)

        if metric == "http_reqs":
            total_reqs += 1
        elif metric == "http_req_failed":
            failed_count += val
        elif metric in ("http_req_duration", "booking_flow_duration_ms"):
            durations[metric].append(val)
        elif metric in ("booking_total", "booking_failed"):
            counts[metric] += val
        elif metric == "booking_success_rate":
            rates[metric].append(val)

    def percentile(lst, p):
        if not lst:
            return 0
        s = sorted(lst)
        idx = math.ceil(p / 100 * len(s)) - 1
        return s[max(0, idx)]

    duration_secs = 1.0
    # Estimasi durasi dari jumlah data points (rough)
    http_dur = durations.get("http_req_duration", [])

    metrics = {
        "http_reqs": {
            "values": {
                "count": total_reqs,
                "rate": total_reqs / max(duration_secs, 1),
            }
        },
        "http_req_failed": {
            "values": {
                "rate": (failed_count / total_reqs) if total_reqs > 0 else 0,
            }
        },
        "http_req_duration": {
            "values": {
                "p(50)": percentile(http_dur, 50),
                "p(95)": percentile(http_dur, 95),
                "p(99)": percentile(http_dur, 99),
            }
        },
        "booking_success_rate": {
            "values": {
                "rate": (
                    sum(rates["booking_success_rate"])
                    / len(rates["booking_success_rate"])
                )
                if rates["booking_success_rate"]
                else 0,
            }
        },
        "booking_total": {"values": {"count": counts.get("booking_total", 0)}},
        "booking_failed": {"values": {"count": counts.get("booking_failed", 0)}},
        "booking_flow_duration_ms": {
            "values": {
                "p(95)": percentile(durations.get("booking_flow_duration_ms", []), 95),
            }
        },
    }

    print(
        f"  WARN: NDJSON tidak menyimpan req/rate akurat. Jalankan ulang test tanpa --out json= untuk hasil terbaik."
    )
    return metrics


def load_monitor_summary(test_name: str, target: str) -> dict:
    """Baca file ringkasan CPU/Mem dari monitor.ps1."""
    filepath = RESULT_DIR / f"{test_name}_{target}_cpu_mem_summary.txt"
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
TARGETS = ["hotelbooking", "roommaster"]
TEST_TYPES = ["load", "spike", "stress"]
TARGET_LABELS = {
    "hotelbooking": "Go (hotelbooking)",
    "roommaster": "Node.js (RoomMasterb)",
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
    sep = "=" * 90
    sep2 = "-" * 90

    print(f"\n{sep}")
    print(f"  TABEL PERBANDINGAN KINERJA: hotelbooking (Go) vs RoomMasterb (Node.js)")
    print(f"{sep}")

    for test in TEST_TYPES:
        print(f"\n  [ {test.upper()} TESTING ]")
        print(f"  {sep2}")
        header = (
            f"  {'Metrik':<30} {'Go (hotelbooking)':>22} {'Node.js (RoomMasterb)':>22}"
        )
        print(header)
        print(f"  {sep2}")

        metrics_list = [
            ("Total HTTP Requests", "total_requests", ""),
            ("Req / detik", "req_per_sec", "rps"),
            ("Error Rate", "error_rate_pct", "%"),
            ("Booking Success Rate", "success_rate_pct", "%"),
            ("Total Booking Sukses", "total_bookings", ""),
            ("Total Booking Gagal", "total_failed", ""),
            ("HTTP Response P50", "http_p50_ms", "ms"),
            ("HTTP Response P95", "http_p95_ms", "ms"),
            ("HTTP Response P99", "http_p99_ms", "ms"),
            ("Flow Duration P95", "flow_p95_ms", "ms"),
            ("CPU Rata-rata", "cpu_avg", ""),
            ("CPU Maksimum", "cpu_max", ""),
            ("Memory Rata-rata", "mem_avg", ""),
            ("Memory Maksimum", "mem_max", ""),
        ]

        for label, key, unit in metrics_list:
            hb_val = results.get("hotelbooking", {}).get(test)
            rm_val = results.get("roommaster", {}).get(test)

            hb_str = (
                f"{hb_val[key]}{unit}"
                if hb_val and hb_val.get(key) is not None
                else "N/A"
            )
            rm_str = (
                f"{rm_val[key]}{unit}"
                if rm_val and rm_val.get(key) is not None
                else "N/A"
            )

            print(f"  {label:<30} {hb_str:>22} {rm_str:>22}")

        print(f"  {sep2}")

    print(f"\n{sep}\n")


print_table()

# Simpan tabel ke file
table_file = RESULT_DIR / "comparison_table.txt"
RESULT_DIR.mkdir(exist_ok=True)
import contextlib
import io

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
X = list(range(len(TEST_TYPES)))
BAR_W = 0.35


CHARTS = [
    # (metric_key,       title,                          xlabel,          ylabel,                    unit,  higher_is_better)
    ("total_requests", "Total HTTP Requests", "Jenis Test", "Jumlah Request", "", True),
    (
        "req_per_sec",
        "Throughput (Req/s)",
        "Jenis Test",
        "Request per Detik (rps)",
        "",
        True,
    ),
    ("error_rate_pct", "Error Rate", "Jenis Test", "Error Rate (%)", "%", False),
    (
        "success_rate_pct",
        "Booking Success Rate",
        "Jenis Test",
        "Success Rate (%)",
        "%",
        True,
    ),
    (
        "total_bookings",
        "Total Booking Sukses",
        "Jenis Test",
        "Jumlah Booking",
        "",
        True,
    ),
    ("total_failed", "Total Booking Gagal", "Jenis Test", "Jumlah Gagal", "", False),
    (
        "http_p50_ms",
        "HTTP Response Time P50",
        "Jenis Test",
        "Waktu Respons (ms)",
        "ms",
        False,
    ),
    (
        "http_p95_ms",
        "HTTP Response Time P95",
        "Jenis Test",
        "Waktu Respons (ms)",
        "ms",
        False,
    ),
    (
        "http_p99_ms",
        "HTTP Response Time P99",
        "Jenis Test",
        "Waktu Respons (ms)",
        "ms",
        False,
    ),
    (
        "flow_p95_ms",
        "Booking Flow Duration P95",
        "Jenis Test",
        "Durasi Flow (ms)",
        "ms",
        False,
    ),
    ("cpu_avg", "CPU Usage — Rata-rata", "Jenis Test", "CPU (%)", "%", False),
    ("cpu_max", "CPU Usage — Maksimum", "Jenis Test", "CPU (%)", "%", False),
    ("mem_avg", "Memory Usage — Rata-rata", "Jenis Test", "Memory (MB)", "MB", False),
    ("mem_max", "Memory Usage — Maksimum", "Jenis Test", "Memory (MB)", "MB", False),
]

# ─── Satu gambar berisi semua subplot ────────────────────────────────────────
NCOLS = 3
NROWS = -(-len(CHARTS) // NCOLS)  # ceil division
fig, axes = plt.subplots(NROWS, NCOLS, figsize=(NCOLS * 6, NROWS * 4.5))
fig.suptitle(
    "Perbandingan Kinerja API: Go (hotelbooking) vs Node.js (RoomMasterb)",
    fontsize=14,
    fontweight="bold",
    y=1.01,
)
axes_flat = axes.flatten()

for idx, (metric_key, title, xlabel, ylabel, unit, higher_is_better) in enumerate(
    CHARTS
):
    ax = axes_flat[idx]

    for i, target in enumerate(TARGETS):
        vals = []
        for test in TEST_TYPES:
            r = results[target].get(test)
            try:
                raw_val = r[metric_key] if r else 0
                v = (
                    float(
                        str(raw_val)
                        .replace("%", "")
                        .replace("ms", "")
                        .replace("MB", "")
                        .strip()
                    )
                    if raw_val not in (None, "N/A", "")
                    else 0
                )
            except (TypeError, ValueError):
                v = 0
            vals.append(v)

        offset = (i - 0.5) * BAR_W
        bars = ax.bar(
            [x + offset for x in X],
            vals,
            BAR_W,
            label=TARGET_LABELS[target],
            color=COLORS[target],
            alpha=0.85,
        )
        for bar, v in zip(bars, vals):
            if v > 0:
                ax.text(
                    bar.get_x() + bar.get_width() / 2,
                    bar.get_height(),
                    f"{v:.0f}{unit}",
                    ha="center",
                    va="bottom",
                    fontsize=6.5,
                )

    ax.set_title(title, fontsize=9, fontweight="bold", pad=6)
    ax.set_xlabel(xlabel, fontsize=7.5, labelpad=4)
    ax.set_ylabel(ylabel, fontsize=7.5, labelpad=4)
    ax.set_xticks(X)
    ax.set_xticklabels([t.capitalize() for t in TEST_TYPES], fontsize=8)
    ax.tick_params(axis="y", labelsize=7)
    ax.grid(axis="y", alpha=0.3, linestyle="--")
    note = "↑ lebih baik" if higher_is_better else "↓ lebih baik"
    ax.annotate(
        note,
        xy=(0.98, 0.97),
        xycoords="axes fraction",
        ha="right",
        va="top",
        fontsize=6,
        color="gray",
        style="italic",
    )

# Sembunyikan subplot kosong (jika jumlah metrik tidak habis dibagi NCOLS)
for idx in range(len(CHARTS), len(axes_flat)):
    axes_flat[idx].set_visible(False)

# Legend bersama di bawah gambar
handles = [plt.Rectangle((0, 0), 1, 1, color=COLORS[t], alpha=0.85) for t in TARGETS]
labels = [TARGET_LABELS[t] for t in TARGETS]
fig.legend(
    handles, labels, loc="lower center", ncol=2, fontsize=9, bbox_to_anchor=(0.5, -0.01)
)

plt.tight_layout()
outpath = RESULT_DIR / "chart_all_metrics.png"
plt.savefig(outpath, dpi=150, bbox_inches="tight")
plt.close()
print(f"Chart disimpan: {outpath}")

print(f"\nSelesai! Semua file ada di: {RESULT_DIR}")
