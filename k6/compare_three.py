"""
compare_three.py — Komparasi 3-arah hasil k6:
    - pure net/http (biru muda)
    - Echo          (biru tua)
    - Node.js       (hijau)

Menghasilkan tabel teks + 3 chart PNG.

Prasyarat — hasil k6 sudah di-backup ke folder berikut:
    k6/results/purehttp/{load,spike,stress}_hotelbooking.json   (+ _cpu_mem_summary.txt)
    k6/results/echo/{load,spike,stress}_hotelbooking.json       (+ _cpu_mem_summary.txt)
    k6/results/{load,spike,stress}_roommaster.json              (+ _cpu_mem_summary.txt)
        (atau di k6/results/nodejs/ kalau di-backup terpisah)

Cara pakai:
    python3 k6/compare_three.py
"""
import json
import math
import sys
from pathlib import Path

RESULT_DIR = Path(__file__).parent / "results"
TESTS = ["load", "spike", "stress"]

# (label, warna, folder, suffix)  — folder relatif ke RESULT_DIR
SOURCES = [
    ("Go net/http", "#42A5F5", "purehttp", "hotelbooking"),  # biru muda
    ("Go echo", "#0D47A1", "echo", "hotelbooking"),          # biru tua
    ("Node.js", "#4CAF50", "nodejs", "roommaster"),          # hijau
]


# ─── Parser hasil k6 ─────────────────────────────────────────────────────────
def parse_k6_result(filepath):
    """Coba format summary JSON dulu, fallback ke NDJSON."""
    if not filepath.exists():
        return None
    raw = filepath.read_text().strip()
    if not raw:
        return None
    try:
        data = json.loads(raw)
        if isinstance(data, dict) and "metrics" in data:
            return _from_summary(data["metrics"])
    except json.JSONDecodeError:
        pass
    return _from_ndjson(raw)


def _from_summary(metrics):
    def v(key, stat):
        return metrics.get(key, {}).get("values", {}).get(stat, 0)

    return {
        "total_reqs": v("http_reqs", "count"),
        "req_per_sec": v("http_reqs", "rate"),
        "error_rate": v("http_req_failed", "rate") * 100,
        "success_rate": v("booking_success_rate", "rate") * 100,
        "p50": v("http_req_duration", "med"),
        "p95": v("http_req_duration", "p(95)"),
        "p99": v("http_req_duration", "p(99)"),
        "flow_p95": v("booking_flow_duration_ms", "p(95)"),
    }


def _from_ndjson(raw):
    durations, flow_durations, succ_rates = [], [], []
    failed = total_reqs = 0
    for line in raw.splitlines():
        line = line.strip()
        if not line:
            continue
        try:
            obj = json.loads(line)
        except json.JSONDecodeError:
            continue
        if not isinstance(obj, dict) or obj.get("type") != "Point":
            continue
        metric = obj.get("metric", "")
        val = obj.get("data", {}).get("value", 0)
        if metric == "http_reqs":
            total_reqs += 1
        elif metric == "http_req_failed":
            failed += val
        elif metric == "http_req_duration":
            durations.append(val)
        elif metric == "booking_flow_duration_ms":
            flow_durations.append(val)
        elif metric == "booking_success_rate":
            succ_rates.append(val)

    def pct(lst, p):
        if not lst:
            return 0
        s = sorted(lst)
        return s[max(0, math.ceil(p / 100 * len(s)) - 1)]

    return {
        "total_reqs": total_reqs,
        "req_per_sec": 0,  # tidak akurat dari NDJSON
        "error_rate": (failed / total_reqs * 100) if total_reqs else 0,
        "success_rate": (sum(succ_rates) / len(succ_rates) * 100) if succ_rates else 0,
        "p50": pct(durations, 50),
        "p95": pct(durations, 95),
        "p99": pct(durations, 99),
        "flow_p95": pct(flow_durations, 95),
    }


def read_monitor(folder, test, suffix):
    """Baca CPU/Mem summary. Fallback: folder utama kalau subfolder tidak ada."""
    candidates = [
        RESULT_DIR / folder / f"{test}_{suffix}_cpu_mem_summary.txt",
        RESULT_DIR / f"{test}_{suffix}_cpu_mem_summary.txt",
    ]
    res = {"cpu_avg": 0.0, "cpu_max": 0.0, "mem_avg": 0.0, "mem_max": 0.0}
    for f in candidates:
        if not f.exists():
            continue
        for line in f.read_text().splitlines():
            for key, tag in (
                ("cpu_avg", "CPU Avg"),
                ("cpu_max", "CPU Max"),
                ("mem_avg", "Mem Avg"),
                ("mem_max", "Mem Max"),
            ):
                if tag in line:
                    num = line.split(":")[-1].strip().replace("%", "").replace("MB", "").strip()
                    try:
                        res[key] = float(num)
                    except ValueError:
                        pass
        return res
    return res


def load_result(folder, test, suffix):
    """Ambil hasil k6. Coba subfolder dulu, fallback ke folder utama."""
    candidates = [
        RESULT_DIR / folder / f"{test}_{suffix}.json",
        RESULT_DIR / f"{test}_{suffix}.json",
    ]
    for f in candidates:
        r = parse_k6_result(f)
        if r:
            r.update(read_monitor(folder, test, suffix))
            return r
    return None


# ─── Kumpulkan semua data ────────────────────────────────────────────────────
# data[test][label] = dict metrik
data = {}
for test in TESTS:
    data[test] = {}
    for label, color, folder, suffix in SOURCES:
        data[test][label] = load_result(folder, test, suffix)


# ─── Tabel teks ──────────────────────────────────────────────────────────────
def fmt(v, unit=""):
    if v is None:
        return "N/A"
    if isinstance(v, float):
        return f"{v:.1f}{unit}" if v < 100 else f"{v:.0f}{unit}"
    return f"{v}{unit}"


METRIC_ROWS = [
    ("Total HTTP Requests", "total_reqs", ""),
    ("Req / detik", "req_per_sec", " rps"),
    ("Error Rate", "error_rate", "%"),
    ("Booking Success Rate", "success_rate", "%"),
    ("HTTP P50", "p50", " ms"),
    ("HTTP P95", "p95", " ms"),
    ("HTTP P99", "p99", " ms"),
    ("Flow Duration P95", "flow_p95", " ms"),
    ("CPU Rata-rata", "cpu_avg", " %"),
    ("CPU Maksimum", "cpu_max", " %"),
    ("Memory Rata-rata", "mem_avg", " MB"),
    ("Memory Maksimum", "mem_max", " MB"),
]

LABELS = [s[0] for s in SOURCES]

lines = []
sep = "=" * 86
lines.append(sep)
lines.append("  KOMPARASI 3-ARAH: pure net/http  vs  Echo  vs  Node.js")
lines.append(sep)
for test in TESTS:
    lines.append(f"\n  [ {test.upper()} TESTING ]")
    lines.append("  " + "-" * 84)
    lines.append(f"  {'Metrik':<24}" + "".join(f"{l:>20}" for l in LABELS))
    lines.append("  " + "-" * 84)
    for label, key, unit in METRIC_ROWS:
        cells = []
        for src_label in LABELS:
            r = data[test].get(src_label)
            cells.append(fmt(r[key], unit) if r else "N/A")
        lines.append(f"  {label:<24}" + "".join(f"{c:>20}" for c in cells))
    lines.append("  " + "-" * 84)
lines.append("\n" + sep)

table = "\n".join(lines)
print(table)
(RESULT_DIR / "comparison_three.txt").write_text(table + "\n")
print(f"\nTabel disimpan: {RESULT_DIR / 'comparison_three.txt'}")


# ─── Chart ───────────────────────────────────────────────────────────────────
try:
    import matplotlib

    matplotlib.use("Agg")
    import matplotlib.pyplot as plt
except ImportError:
    print("\nmatplotlib tidak terinstall — chart dilewati. (pip install matplotlib)")
    sys.exit(0)

X = list(range(len(TESTS)))
BAR_W = 0.26


def get_val(test, label, key):
    r = data[test].get(label)
    if not r or r.get(key) is None:
        return 0.0
    try:
        return float(r[key])
    except (TypeError, ValueError):
        return 0.0


def draw_pair(filename, left, right):
    """1 gambar = 2 subplot, tiap grup test punya 3 bar (pure/echo/node)."""
    fig, axes = plt.subplots(1, 2, figsize=(14, 5.5))

    for ax, spec in zip(axes, [left, right]):
        max_val = 0.0
        for i, (label, color, _, _) in enumerate(SOURCES):
            vals = [get_val(test, label, spec["metric_key"]) for test in TESTS]
            max_val = max(max_val, *vals)
            offset = (i - 1) * BAR_W  # 3 bar: -BAR_W, 0, +BAR_W
            bars = ax.bar(
                [x + offset for x in X],
                vals,
                BAR_W,
                label=label,
                color=color,
                alpha=0.9,
            )
            for bar, v in zip(bars, vals):
                if v > 0:
                    ax.text(
                        bar.get_x() + bar.get_width() / 2,
                        bar.get_height(),
                        f"{v:.0f}{spec['unit']}",
                        ha="center",
                        va="bottom",
                        fontsize=8,
                    )

        if max_val > 0:
            ax.set_ylim(0, max_val * 1.20)
        ax.set_title(spec["title"], fontsize=13, fontweight="bold", pad=8)
        ax.set_xlabel("Test Type", fontsize=10, labelpad=6)
        ax.set_ylabel(spec["ylabel"], fontsize=10, labelpad=6)
        ax.set_xticks(X)
        ax.set_xticklabels([t.capitalize() for t in TESTS], fontsize=10)
        ax.tick_params(axis="y", labelsize=9)
        ax.grid(axis="y", alpha=0.3, linestyle="--")
        note = "↑ better" if spec["higher_is_better"] else "↓ better"
        ax.annotate(
            note,
            xy=(0.98, 0.97),
            xycoords="axes fraction",
            ha="right",
            va="top",
            fontsize=9,
            color="gray",
            style="italic",
        )

    handles = [
        plt.Rectangle((0, 0), 1, 1, color=c, alpha=0.9) for _, c, _, _ in SOURCES
    ]
    fig.legend(
        handles,
        LABELS,
        loc="lower center",
        ncol=3,
        fontsize=10,
        bbox_to_anchor=(0.5, -0.02),
    )

    plt.tight_layout(rect=[0, 0.07, 1, 1])
    outpath = RESULT_DIR / filename
    plt.savefig(outpath, dpi=150, bbox_inches="tight")
    plt.close()
    print(f"Chart disimpan: {outpath}")


# Gambar 1: Resource (CPU max + Memory avg)
draw_pair(
    "chart3_resource.png",
    left={
        "metric_key": "cpu_max",
        "title": "CPU Usage (maximum)",
        "ylabel": "CPU (%)",
        "unit": "%",
        "higher_is_better": False,
    },
    right={
        "metric_key": "mem_avg",
        "title": "Memory Usage (average)",
        "ylabel": "Memory (MB)",
        "unit": "MB",
        "higher_is_better": False,
    },
)

# Gambar 2: Latency (HTTP p95 + Flow p95)
draw_pair(
    "chart3_latency.png",
    left={
        "metric_key": "p95",
        "title": "HTTP Response Time (p95)",
        "ylabel": "Response Time (ms)",
        "unit": "ms",
        "higher_is_better": False,
    },
    right={
        "metric_key": "flow_p95",
        "title": "Booking Flow Duration (p95)",
        "ylabel": "Flow Duration (ms)",
        "unit": "ms",
        "higher_is_better": False,
    },
)

# Gambar 3: Throughput (Total Requests + Req/s)
draw_pair(
    "chart3_throughput.png",
    left={
        "metric_key": "total_reqs",
        "title": "Total HTTP Requests",
        "ylabel": "Number of Requests",
        "unit": "",
        "higher_is_better": True,
    },
    right={
        "metric_key": "req_per_sec",
        "title": "Throughput",
        "ylabel": "Requests per Second",
        "unit": "",
        "higher_is_better": True,
    },
)

print(f"\nSelesai! Semua file ada di: {RESULT_DIR}")
