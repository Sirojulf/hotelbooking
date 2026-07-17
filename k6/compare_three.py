"""Compare k6 results for Go net/http, Go Echo, and Node.js.

The charts follow common IEEE figure conventions: two-column width, compact
serif typography, grayscale-safe bar patterns, and vector PDF output. The
existing series colors are preserved exactly. A 600-dpi PNG copy is generated
for workflows that require raster images.

Expected input:
    results/purehttp/{load,spike,stress}_hotelbooking.json
    results/echo/{load,spike,stress}_hotelbooking.json
    results/{load,spike,stress}_roommaster.json

Usage:
    python3 compare_three.py

Output:
    results/comparison_three.txt
    results/chart3_resource.{pdf,png}
    results/chart3_latency.{pdf,png}
    results/chart3_throughput.{pdf,png}
"""
import json
import math
import sys
from pathlib import Path

RESULT_DIR = Path(__file__).parent / "results"
TESTS = ["load", "spike", "stress"]

# Label, color, folder relative to RESULT_DIR, and file suffix.
SOURCES = [
    ("Go net/http", "#42A5F5", "purehttp", "hotelbooking"),
    ("Go Echo", "#0D47A1", "echo", "hotelbooking"),
    ("Node.js", "#4CAF50", "nodejs", "roommaster"),
]


# Parse k6 results.
def parse_k6_result(filepath):
    """Read handleSummary() JSON, falling back to legacy NDJSON."""
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
        "req_per_sec": 0,  # Request rate cannot be recovered from NDJSON.
        "error_rate": (failed / total_reqs * 100) if total_reqs else 0,
        "success_rate": (sum(succ_rates) / len(succ_rates) * 100) if succ_rates else 0,
        "p50": pct(durations, 50),
        "p95": pct(durations, 95),
        "p99": pct(durations, 99),
        "flow_p95": pct(flow_durations, 95),
    }


def read_monitor(folder, test, suffix):
    """Read CPU/memory data, falling back to the main results directory."""
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
                    num = (
                        line.split(":")[-1]
                        .strip()
                        .replace("%", "")
                        .replace("MB", "")
                        .strip()
                    )
                    try:
                        res[key] = float(num)
                    except ValueError:
                        pass
        return res
    return res


def load_result(folder, test, suffix):
    """Load a k6 result from its subfolder or the main results directory."""
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


# Collect all available data as data[test][label].
data = {}
for test in TESTS:
    data[test] = {}
    for label, color, folder, suffix in SOURCES:
        data[test][label] = load_result(folder, test, suffix)


# Build the text report.
def fmt(v, unit=""):
    if v is None:
        return "N/A"
    if isinstance(v, float):
        return f"{v:.1f}{unit}" if v < 100 else f"{v:.0f}{unit}"
    return f"{v}{unit}"


METRIC_ROWS = [
    ("Total HTTP Requests", "total_reqs", ""),
    ("Requests per Second", "req_per_sec", " rps"),
    ("Error Rate", "error_rate", " %"),
    ("Booking Success Rate", "success_rate", " %"),
    ("HTTP P50", "p50", " ms"),
    ("HTTP P95", "p95", " ms"),
    ("HTTP P99", "p99", " ms"),
    ("Flow Duration P95", "flow_p95", " ms"),
    ("Average CPU Usage", "cpu_avg", " %"),
    ("Maximum CPU Usage", "cpu_max", " %"),
    ("Average Memory Usage", "mem_avg", " MB"),
    ("Maximum Memory Usage", "mem_max", " MB"),
]

LABELS = [s[0] for s in SOURCES]

lines = []
sep = "=" * 86
lines.append(sep)
lines.append("  THREE-WAY COMPARISON: Go net/http vs Go Echo vs Node.js")
lines.append(sep)
for test in TESTS:
    lines.append(f"\n  [ {test.upper()} TESTING ]")
    lines.append("  " + "-" * 84)
    lines.append(f"  {'Metric':<24}" + "".join(f"{l:>20}" for l in LABELS))
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
print(f"\nTable saved to: {RESULT_DIR / 'comparison_three.txt'}")


# Create the figures.
try:
    import matplotlib

    matplotlib.use("Agg")
    import matplotlib.pyplot as plt
except ImportError:
    print("\nmatplotlib is not installed; charts were skipped. (pip install matplotlib)")
    sys.exit(0)

X = list(range(len(TESTS)))
BAR_W = 0.25
IEEE_DOUBLE_COLUMN_WIDTH = 7.16
IEEE_FIGURE_HEIGHT = 2.85
PNG_DPI = 600
HATCHES = ["////", "\\\\\\\\", "...."]

plt.rcParams.update(
    {
        "font.family": "serif",
        "font.serif": [
            "Liberation Serif",
            "Times New Roman",
            "Times",
            "DejaVu Serif",
        ],
        "font.size": 8,
        "axes.titlesize": 8,
        "axes.labelsize": 8,
        "xtick.labelsize": 7,
        "ytick.labelsize": 7,
        "legend.fontsize": 7,
        "axes.linewidth": 0.6,
        "lines.linewidth": 0.8,
        "patch.linewidth": 0.6,
        "figure.facecolor": "white",
        "axes.facecolor": "white",
        "savefig.facecolor": "white",
        "pdf.fonttype": 42,
        "ps.fonttype": 42,
        "mathtext.fontset": "stix",
    }
)


def get_val(test, label, key):
    r = data[test].get(label)
    if not r or r.get(key) is None:
        return 0.0
    try:
        return float(r[key])
    except (TypeError, ValueError):
        return 0.0


def format_value(value, style):
    """Format compact bar labels without repeating the axis unit."""
    if style == "decimal":
        return f"{value:.1f}"
    if style == "integer":
        return f"{value:,.0f}"
    if abs(value) >= 1000:
        return f"{value / 1000:.1f}k"
    return f"{value:.0f}"


def draw_pair(basename, left, right):
    """Create an IEEE two-column figure containing two comparison panels."""
    fig, axes = plt.subplots(
        1,
        2,
        figsize=(IEEE_DOUBLE_COLUMN_WIDTH, IEEE_FIGURE_HEIGHT),
        sharex=True,
    )

    for panel_index, (ax, spec) in enumerate(zip(axes, [left, right])):
        max_val = 0.0
        for i, (label, color, _, _) in enumerate(SOURCES):
            vals = [get_val(test, label, spec["metric_key"]) for test in TESTS]
            max_val = max(max_val, *vals)
            offset = (i - 1) * BAR_W
            bars = ax.bar(
                [x + offset for x in X],
                vals,
                BAR_W,
                label=label,
                color=color,
                edgecolor="black",
                linewidth=0.45,
                hatch=HATCHES[i],
                zorder=3,
            )
            for bar, v in zip(bars, vals):
                if v > 0:
                    ax.text(
                        bar.get_x() + bar.get_width() / 2,
                        bar.get_height(),
                        format_value(v, spec.get("value_format", "auto")),
                        ha="center",
                        va="bottom",
                        fontsize=5.8,
                        clip_on=False,
                    )

        if max_val > 0:
            ax.set_ylim(0, max_val * 1.23)

        panel_letter = chr(ord("a") + panel_index)
        ax.set_title(f"({panel_letter}) {spec['title']}", fontweight="normal", pad=4)
        ax.set_xlabel("Workload")
        ax.set_ylabel(spec["ylabel"])
        ax.set_xticks(X)
        ax.set_xticklabels([t.capitalize() for t in TESTS])
        ax.tick_params(axis="both", direction="out", length=2.5, width=0.6)
        ax.grid(axis="y", color="#BFBFBF", linewidth=0.45, linestyle=":", zorder=0)
        ax.spines["top"].set_visible(False)
        ax.spines["right"].set_visible(False)

    handles = [
        plt.Rectangle(
            (0, 0),
            1,
            1,
            facecolor=color,
            edgecolor="black",
            linewidth=0.45,
            hatch=HATCHES[index],
        )
        for index, (_, color, _, _) in enumerate(SOURCES)
    ]
    fig.legend(
        handles,
        LABELS,
        loc="upper center",
        ncol=3,
        frameon=False,
        handlelength=1.8,
        columnspacing=1.2,
        bbox_to_anchor=(0.5, 0.995),
    )

    fig.subplots_adjust(left=0.09, right=0.99, bottom=0.19, top=0.80, wspace=0.28)
    output_paths = []
    for extension, dpi in (("pdf", None), ("png", PNG_DPI)):
        outpath = RESULT_DIR / f"{basename}.{extension}"
        fig.savefig(outpath, dpi=dpi, metadata={"Creator": "compare_three.py"})
        output_paths.append(outpath)
    plt.close()
    print(f"Figures saved: {output_paths[0]} and {output_paths[1]}")


# Figure 1: maximum CPU and average memory usage.
draw_pair(
    "chart3_resource",
    left={
        "metric_key": "cpu_max",
        "title": "CPU Usage (maximum)",
        "ylabel": "CPU (%)",
        "value_format": "decimal",
    },
    right={
        "metric_key": "mem_avg",
        "title": "Memory Usage (average)",
        "ylabel": "Memory (MB)",
        "value_format": "integer",
    },
)

# Figure 2: HTTP and booking-flow latency.
draw_pair(
    "chart3_latency",
    left={
        "metric_key": "p95",
        "title": "HTTP P95 Latency",
        "ylabel": "Response Time (ms)",
        "value_format": "integer",
    },
    right={
        "metric_key": "flow_p95",
        "title": "Booking-Flow P95 Latency",
        "ylabel": "Flow Duration (ms)",
        "value_format": "integer",
    },
)

# Figure 3: total requests and throughput.
draw_pair(
    "chart3_throughput",
    left={
        "metric_key": "total_reqs",
        "title": "Total HTTP Requests",
        "ylabel": "Number of Requests",
        "value_format": "compact",
    },
    right={
        "metric_key": "req_per_sec",
        "title": "Throughput",
        "ylabel": "Throughput (requests/s)",
        "value_format": "decimal",
    },
)

print(f"\nDone. All output files are in: {RESULT_DIR}")
