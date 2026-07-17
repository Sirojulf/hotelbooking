"""Analyze per-endpoint latency in the booking flow.

The figures use IEEE two-column dimensions, compact serif typography,
grayscale-safe patterns, vector PDF output, and 600-dpi PNG output.

Run after:
    k6 run endpoint_breakdown.js

Usage:
    python3 analyze_endpoints.py

Output:
    results/endpoint_breakdown_table.txt
    results/chart_endpoint_breakdown.{pdf,png}
    results/chart_endpoint_contribution.{pdf,png}
"""
import json
import sys
from pathlib import Path

RESULT_DIR = Path(__file__).parent / "results"

ENDPOINTS = [
    ("check_availability", "GET /rooms/{id}/availability"),
    ("create_reservation", "POST /guests/reservations"),
    ("pay_reservation", "POST /guests/reservations/{id}/pay"),
    ("cancel_reservation", "POST /guests/reservations/{id}/cancel"),
]

# Available files are loaded in this display order.
TARGETS = [
    ("Go net/http", "endpoint_breakdown_purehttp.json"),
    ("Go Echo", "endpoint_breakdown_hotelbooking.json"),
    ("Node.js", "endpoint_breakdown_roommaster.json"),
]


# Parse per-endpoint metrics from k6 summary JSON.
def parse(filepath):
    if not filepath.exists():
        return None
    try:
        data = json.loads(filepath.read_text())
    except json.JSONDecodeError:
        return None
    m = data.get("metrics", {})

    def stat(key, statname):
        return m.get(key, {}).get("values", {}).get(statname, 0)

    out = {}
    for ep, _ in ENDPOINTS:
        out[ep] = {
            "avg": stat(f"ep_{ep}_ms", "avg"),
            "med": stat(f"ep_{ep}_ms", "med"),
            "p90": stat(f"ep_{ep}_ms", "p(90)"),
            "p95": stat(f"ep_{ep}_ms", "p(95)"),
            "p99": stat(f"ep_{ep}_ms", "p(99)"),
            "max": stat(f"ep_{ep}_ms", "max"),
            "ok": stat(f"ep_{ep}_ok", "rate") * 100,
        }
    return out


# Load all available targets.
results = {}
for label, fname in TARGETS:
    r = parse(RESULT_DIR / fname)
    if r:
        results[label] = r

if not results:
    print("ERROR: no endpoint_breakdown_*.json files were found in results/.")
    print("Run this first: k6 run endpoint_breakdown.js")
    sys.exit(1)


# Build the text report.
def format_ms(value):
    return f"{value:.0f} ms"


lines = []
sep = "=" * 88
lines.append(sep)
lines.append("  ENDPOINT BREAKDOWN: DIRECT BOOKING FLOW")
lines.append(sep)

for target, ep_data in results.items():
    lines.append(f"\n  Target: {target}")
    lines.append("  " + "-" * 86)
    header = f"  {'Endpoint':<22}{'Avg':>10}{'Med':>10}{'P90':>10}{'P95':>10}{'P99':>10}{'Max':>10}{'OK%':>8}"
    lines.append(header)
    lines.append("  " + "-" * 86)

    total_p95 = sum(ep_data[ep]["p95"] for ep, _ in ENDPOINTS)
    for ep, _ in ENDPOINTS:
        d = ep_data[ep]
        lines.append(
            f"  {ep:<22}"
            + f"{format_ms(d['avg']):>10}"
            + f"{format_ms(d['med']):>10}"
            + f"{format_ms(d['p90']):>10}"
            + f"{format_ms(d['p95']):>10}"
            + f"{format_ms(d['p99']):>10}"
            + f"{format_ms(d['max']):>10}"
            + f"{d['ok']:>7.1f}%"
        )
    lines.append("  " + "-" * 86)
    heaviest = max(ENDPOINTS, key=lambda x: ep_data[x[0]]["p95"])
    contribution = (
        ep_data[heaviest[0]]["p95"] / total_p95 * 100 if total_p95 else 0
    )
    lines.append(
        f"  Heaviest endpoint (P95): {heaviest[0]} "
        f"({ep_data[heaviest[0]]['p95']:.0f} ms, "
        f"{contribution:.1f}% of total flow)"
    )

lines.append("\n" + sep)
table = "\n".join(lines)
print(table)
(RESULT_DIR / "endpoint_breakdown_table.txt").write_text(table + "\n")
print(f"\nTable saved to: {RESULT_DIR / 'endpoint_breakdown_table.txt'}")


# Create the figures.
try:
    import matplotlib
    matplotlib.use("Agg")
    import matplotlib.pyplot as plt
except ImportError:
    print("\nmatplotlib is not installed; charts were skipped.")
    sys.exit(0)

import numpy as np

EP_LABELS = [ep for ep, _ in ENDPOINTS]
EP_DISPLAY_LABELS = [
    "Check\navailability",
    "Create\nreservation",
    "Pay\nreservation",
    "Cancel\nreservation",
]
TARGET_COLORS = {
    "Go Echo": "#0D47A1",
    "Go net/http": "#42A5F5",
    "Node.js": "#4CAF50",
}
TARGET_HATCHES = {
    "Go net/http": "////",
    "Go Echo": "\\\\\\\\",
    "Node.js": "....",
}
ENDPOINT_COLORS = {
    "check_availability": "#FFC107",
    "create_reservation": "#E53935",
    "pay_reservation": "#FB8C00",
    "cancel_reservation": "#1E88E5",
}
ENDPOINT_HATCHES = ["////", "\\\\\\\\", "....", "xxxx"]
ENDPOINT_LEGEND_LABELS = ["Availability", "Creation", "Payment", "Cancellation"]
IEEE_DOUBLE_COLUMN_WIDTH = 7.16
IEEE_FIGURE_HEIGHT = 2.85
PNG_DPI = 600

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


def style_axis(ax):
    """Apply shared publication styling to an axis."""
    ax.tick_params(axis="both", direction="out", length=2.5, width=0.6)
    ax.grid(axis="y", color="#BFBFBF", linewidth=0.45, linestyle=":", zorder=0)
    ax.spines["top"].set_visible(False)
    ax.spines["right"].set_visible(False)


def save_figure(fig, basename):
    """Save a vector PDF and a 600-dpi PNG with identical dimensions."""
    output_paths = []
    for extension, dpi in (("pdf", None), ("png", PNG_DPI)):
        outpath = RESULT_DIR / f"{basename}.{extension}"
        fig.savefig(outpath, dpi=dpi, metadata={"Creator": "analyze_endpoints.py"})
        output_paths.append(outpath)
    plt.close(fig)
    print(f"Figures saved: {output_paths[0]} and {output_paths[1]}")


# Figure 1: grouped P95 latency by endpoint and target.
def chart_p95():
    targets = list(results.keys())
    n = len(targets)
    x = np.arange(len(EP_LABELS))
    width = 0.8 / max(n, 1)

    fig, ax = plt.subplots(
        figsize=(IEEE_DOUBLE_COLUMN_WIDTH, IEEE_FIGURE_HEIGHT)
    )
    max_val = 0.0
    for i, target in enumerate(targets):
        vals = [results[target][ep]["p95"] for ep in EP_LABELS]
        max_val = max(max_val, *vals)
        offset = (i - (n - 1) / 2) * width
        bars = ax.bar(
            x + offset,
            vals,
            width,
            label=target,
            color=TARGET_COLORS.get(target, "#888888"),
            edgecolor="black",
            linewidth=0.45,
            hatch=TARGET_HATCHES[target],
            zorder=3,
        )
        for bar, v in zip(bars, vals):
            if v > 0:
                ax.text(
                    bar.get_x() + bar.get_width() / 2,
                    bar.get_height(),
                    f"{v:.0f}",
                    ha="center",
                    va="bottom",
                    fontsize=6,
                    clip_on=False,
                )

    ax.set_ylim(0, max_val * 1.22 if max_val else 1)
    ax.set_title("Endpoint P95 Latency", fontweight="normal", pad=4)
    ax.set_ylabel("P95 Latency (ms)")
    ax.set_xlabel("Booking-Flow Endpoint")
    ax.set_xticks(x)
    ax.set_xticklabels(EP_DISPLAY_LABELS)
    style_axis(ax)
    ax.legend(
        loc="upper center",
        bbox_to_anchor=(0.5, 1.24),
        ncol=3,
        frameon=False,
        handlelength=1.8,
        columnspacing=1.2,
    )
    fig.subplots_adjust(left=0.09, right=0.99, bottom=0.24, top=0.76)
    save_figure(fig, "chart_endpoint_breakdown")


# Figure 2: stacked endpoint contribution to total P95 flow latency.
def chart_contribution():
    targets = list(results.keys())
    fig, ax = plt.subplots(
        figsize=(IEEE_DOUBLE_COLUMN_WIDTH, IEEE_FIGURE_HEIGHT)
    )

    bottoms = np.zeros(len(targets))
    for endpoint_index, ep in enumerate(EP_LABELS):
        vals = np.array([results[t][ep]["p95"] for t in targets])
        bars = ax.bar(
            targets,
            vals,
            bottom=bottoms,
            width=0.52,
            label=ENDPOINT_LEGEND_LABELS[endpoint_index],
            color=ENDPOINT_COLORS[ep],
            edgecolor="black",
            linewidth=0.45,
            hatch=ENDPOINT_HATCHES[endpoint_index],
            zorder=3,
        )
        for bar, bottom, value in zip(bars, bottoms, vals):
            if value > 30:
                ax.text(
                    bar.get_x() + bar.get_width() / 2,
                    bottom + value / 2,
                    f"{value:.0f}",
                    ha="center",
                    va="center",
                    fontsize=6.2,
                    color="black",
                )
        bottoms += vals

    for i, total in enumerate(bottoms):
        ax.text(
            i,
            total + max(bottoms) * 0.025,
            f"{total:.0f} ms",
            ha="center",
            va="bottom",
            fontsize=6.5,
        )

    ax.set_ylim(0, max(bottoms) * 1.17 if len(bottoms) else 1)
    ax.set_title("Endpoint Contribution to P95 Flow Latency", fontweight="normal", pad=4)
    ax.set_ylabel("Cumulative P95 Latency (ms)")
    ax.set_xlabel("Implementation")
    style_axis(ax)
    ax.legend(
        loc="upper center",
        bbox_to_anchor=(0.5, 1.24),
        ncol=4,
        frameon=False,
        handlelength=1.8,
        columnspacing=1.0,
    )
    fig.subplots_adjust(left=0.09, right=0.99, bottom=0.19, top=0.76)
    save_figure(fig, "chart_endpoint_contribution")


chart_p95()
chart_contribution()
print(f"\nDone. All output files are in: {RESULT_DIR}")
