"""Compare k6 results for hotelbooking (Go) and RoomMaster (Node.js).

The charts follow common IEEE figure conventions: two-column width, compact
serif typography, grayscale-safe bar patterns, and vector PDF output. A
600-dpi PNG copy is also generated for workflows that require raster images.

Usage:
    python3 analyze.py

Dependency:
    pip install matplotlib

Output:
    results/comparison_table.txt
    results/chart_resource.{pdf,png}
    results/chart_latency.{pdf,png}
    results/chart_throughput.{pdf,png}
"""

import contextlib
import io
import json
import sys
from pathlib import Path

RESULT_DIR = Path(__file__).parent / "results"

# Matplotlib is optional so the text report can still be generated on servers.
try:
    import matplotlib

    matplotlib.use("Agg")  # non-interactive backend
    import matplotlib.pyplot as plt

    HAS_MATPLOTLIB = True
except ImportError:
    HAS_MATPLOTLIB = False
    print("INFO: matplotlib is not installed; only the text table will be created.")
    print("      Install it with: pip install matplotlib")


# Load k6 JSON results.
def load_k6_result(filepath: Path) -> dict:
    """Read a k6 result and extract the metrics used in the report.

    Supported formats:
      - Summary JSON produced by handleSummary()
      - NDJSON produced by --out json= (legacy fallback)
    """
    if not filepath.exists():
        return None

    raw = filepath.read_text(encoding="utf-8").strip()
    if not raw:
        return None

    # Try the handleSummary() JSON format first.
    try:
        data = json.loads(raw)
        if "metrics" not in data:
            print(f"WARNING: {filepath.name} has no 'metrics' key; skipping it.")
            return None
        metrics = data["metrics"]
    except json.JSONDecodeError:
        print(f"INFO: detected NDJSON in {filepath.name}; parsing the data points.")
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


def _parse_ndjson(raw: str, _filepath: Path) -> dict | None:
    """Convert k6 NDJSON data points into a small summary structure."""
    import math
    from collections import defaultdict

    counts = defaultdict(float)
    rates = defaultdict(list)
    durations = defaultdict(list)
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
                "med": percentile(http_dur, 50),
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
        "  WARNING: request rate cannot be reconstructed accurately from this "
        "NDJSON file. Run the test again with handleSummary() for an exact value."
    )
    return metrics


def load_monitor_summary(test_name: str, target: str) -> dict:
    """Read CPU and memory metrics produced by the monitoring script."""
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


# Collect all available results.
TARGETS = ["hotelbooking", "roommaster"]
TEST_TYPES = ["load", "spike", "stress"]
TARGET_LABELS = {
    "hotelbooking": "Go (hotelbooking)",
    "roommaster": "Node.js (RoomMaster)",
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


# Print the text report.
def print_table():
    sep = "=" * 90
    sep2 = "-" * 90

    print(f"\n{sep}")
    print("  PERFORMANCE COMPARISON: hotelbooking (Go) vs RoomMaster (Node.js)")
    print(f"{sep}")

    for test in TEST_TYPES:
        print(f"\n  [ {test.upper()} TESTING ]")
        print(f"  {sep2}")
        header = (
            f"  {'Metric':<30} {'Go (hotelbooking)':>22} {'Node.js (RoomMaster)':>22}"
        )
        print(header)
        print(f"  {sep2}")

        metrics_list = [
            ("Total HTTP Requests", "total_requests", ""),
            ("Requests per Second", "req_per_sec", " rps"),
            ("Error Rate", "error_rate_pct", " %"),
            ("Booking Success Rate", "success_rate_pct", " %"),
            ("Successful Bookings", "total_bookings", ""),
            ("Failed Bookings", "total_failed", ""),
            ("HTTP Response P50", "http_p50_ms", " ms"),
            ("HTTP Response P95", "http_p95_ms", " ms"),
            ("HTTP Response P99", "http_p99_ms", " ms"),
            ("Flow Duration P95", "flow_p95_ms", " ms"),
            ("Average CPU Usage", "cpu_avg", ""),
            ("Maximum CPU Usage", "cpu_max", ""),
            ("Average Memory Usage", "mem_avg", ""),
            ("Maximum Memory Usage", "mem_max", ""),
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

# Save the text report.
table_file = RESULT_DIR / "comparison_table.txt"
RESULT_DIR.mkdir(exist_ok=True)
buf = io.StringIO()
with contextlib.redirect_stdout(buf):
    print_table()
table_file.write_text(buf.getvalue())
print(f"Table saved to: {table_file}")

# Create the figures.
if not HAS_MATPLOTLIB:
    print("Done (charts were not generated).")
    sys.exit(0)

IEEE_DOUBLE_COLUMN_WIDTH = 7.16
IEEE_FIGURE_HEIGHT = 2.85
PNG_DPI = 600

# Colorblind-safe colors plus distinct hatches preserve meaning in grayscale.
COLORS = {"hotelbooking": "#0072B2", "roommaster": "#D55E00"}
HATCHES = {"hotelbooking": "////", "roommaster": "\\\\\\\\"}
CHART_LABELS = TARGET_LABELS
X = list(range(len(TEST_TYPES)))
BAR_W = 0.34

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


def get_val(target, test, metric_key):
    """Return a metric as a float after removing units from monitor data."""
    r = results[target].get(test)
    try:
        raw_val = r[metric_key] if r else 0
        if raw_val in (None, "N/A", ""):
            return 0.0
        return float(
            str(raw_val).replace("%", "").replace("ms", "").replace("MB", "").strip()
        )
    except (TypeError, ValueError):
        return 0.0


def format_value(value: float, style: str) -> str:
    """Format compact bar labels without repeating the axis unit."""
    if style == "decimal":
        return f"{value:.1f}"
    if style == "integer":
        return f"{value:,.0f}"
    if abs(value) >= 1000:
        return f"{value / 1000:.1f}k"
    return f"{value:.0f}"


def draw_pair(basename, left, right):
    """Create a publication-ready IEEE two-column figure with two panels.

    ``left`` and ``right`` define metric_key, title, ylabel, and value_format.
    The PDF is the preferred LaTeX asset; PNG is provided as a raster fallback.
    """
    fig, axes = plt.subplots(
        1,
        2,
        figsize=(IEEE_DOUBLE_COLUMN_WIDTH, IEEE_FIGURE_HEIGHT),
        sharex=True,
    )

    for panel_index, (ax, spec) in enumerate(zip(axes, [left, right])):
        max_val = 0.0
        for i, target in enumerate(TARGETS):
            vals = [get_val(target, test, spec["metric_key"]) for test in TEST_TYPES]
            max_val = max(max_val, *vals)
            offset = (i - 0.5) * BAR_W
            bars = ax.bar(
                [x + offset for x in X],
                vals,
                BAR_W,
                label=CHART_LABELS[target],
                color=COLORS[target],
                edgecolor="black",
                linewidth=0.5,
                hatch=HATCHES[target],
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
                        fontsize=6.5,
                        rotation=0,
                        clip_on=False,
                    )

        # Reserve enough headroom for value labels at final publication size.
        if max_val > 0:
            ax.set_ylim(0, max_val * 1.22)

        panel_letter = chr(ord("a") + panel_index)
        ax.set_title(f"({panel_letter}) {spec['title']}", fontweight="normal", pad=4)
        ax.set_xlabel("Workload")
        ax.set_ylabel(spec["ylabel"])
        ax.set_xticks(X)
        ax.set_xticklabels([t.capitalize() for t in TEST_TYPES])
        ax.tick_params(axis="both", direction="out", length=2.5, width=0.6)
        ax.grid(axis="y", color="#BFBFBF", linewidth=0.45, linestyle=":", zorder=0)
        ax.spines["top"].set_visible(False)
        ax.spines["right"].set_visible(False)

    handles = [
        plt.Rectangle(
            (0, 0),
            1,
            1,
            facecolor=COLORS[target],
            edgecolor="black",
            linewidth=0.5,
            hatch=HATCHES[target],
        )
        for target in TARGETS
    ]
    labels = [CHART_LABELS[t] for t in TARGETS]
    fig.legend(
        handles,
        labels,
        loc="upper center",
        ncol=2,
        frameon=False,
        handlelength=1.8,
        columnspacing=1.5,
        bbox_to_anchor=(0.5, 0.995),
    )

    fig.subplots_adjust(left=0.09, right=0.99, bottom=0.19, top=0.80, wspace=0.28)
    output_paths = []
    for extension, dpi in (("pdf", None), ("png", PNG_DPI)):
        outpath = RESULT_DIR / f"{basename}.{extension}"
        fig.savefig(outpath, dpi=dpi, metadata={"Creator": "analyze.py"})
        output_paths.append(outpath)
    plt.close()
    print(f"Figures saved: {output_paths[0]} and {output_paths[1]}")


# Figure 1: resource usage (maximum CPU and average memory).
draw_pair(
    "chart_resource",
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
    "chart_latency",
    left={
        "metric_key": "http_p95_ms",
        "title": "HTTP P95 Latency",
        "ylabel": "Response Time (ms)",
        "value_format": "integer",
    },
    right={
        "metric_key": "flow_p95_ms",
        "title": "Booking-Flow P95 Latency",
        "ylabel": "Flow Duration (ms)",
        "value_format": "integer",
    },
)

# Figure 3: total requests and throughput.
draw_pair(
    "chart_throughput",
    left={
        "metric_key": "total_requests",
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
