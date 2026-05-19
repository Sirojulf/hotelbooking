"""
analyze_endpoints.py — Analisis per-endpoint breakdown booking flow.
Untuk paper: identifikasi endpoint mana yang paling berat.

Jalankan SETELAH:
    k6 run k6/endpoint_breakdown.js
    (atau dengan -e TARGET=roommaster untuk Node.js)

Cara pakai:
    python3 k6/analyze_endpoints.py

Output:
    k6/results/endpoint_breakdown_table.txt
    k6/results/chart_endpoint_breakdown.png   (grouped bar: tiap endpoint, p50/p95/p99)
    k6/results/chart_endpoint_contribution.png (stacked: kontribusi tiap endpoint ke total flow)
"""
import json
import sys
from pathlib import Path

RESULT_DIR = Path(__file__).parent / "results"

ENDPOINTS = [
    ("check_availability", "GET /rooms/{id}/availability"),
    ("create_reservation", "POST /guests/reservations"),
    ("pay_reservation",    "POST /guests/reservations/{id}/pay"),
    ("cancel_reservation", "POST /guests/reservations/{id}/cancel"),
]

# Target list — yang ada datanya akan dimuat. Urutan = urutan tampil di chart.
# Filename match `endpoint_breakdown_${LABEL}.json` dari k6 script.
TARGETS = [
    ("Go net/http", "endpoint_breakdown_purehttp.json"),
    ("Go echo",     "endpoint_breakdown_hotelbooking.json"),
    ("Node.js",     "endpoint_breakdown_roommaster.json"),
]


# ─── Parse k6 summary JSON → per-endpoint stats ──────────────────────────────
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
            "ok":  stat(f"ep_{ep}_ok", "rate") * 100,
        }
    return out


# ─── Load semua data ─────────────────────────────────────────────────────────
results = {}
for label, fname in TARGETS:
    r = parse(RESULT_DIR / fname)
    if r:
        results[label] = r

if not results:
    print("ERROR: tidak ada hasil endpoint_breakdown_*.json di k6/results/")
    print("Jalankan dulu: k6 run k6/endpoint_breakdown.js")
    sys.exit(1)


# ─── Cetak tabel teks ────────────────────────────────────────────────────────
lines = []
sep = "=" * 88
lines.append(sep)
lines.append("  ENDPOINT BREAKDOWN — direct booking flow")
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
        contrib = (d["p95"] / total_p95 * 100) if total_p95 else 0
        lines.append(
            f"  {ep:<22}" +
            f"{d['avg']:>8.0f}ms" +
            f"{d['med']:>8.0f}ms" +
            f"{d['p90']:>8.0f}ms" +
            f"{d['p95']:>8.0f}ms" +
            f"{d['p99']:>8.0f}ms" +
            f"{d['max']:>8.0f}ms" +
            f"{d['ok']:>7.1f}%"
        )
    lines.append("  " + "-" * 86)
    # Highlight endpoint terberat (by p95)
    heaviest = max(ENDPOINTS, key=lambda x: ep_data[x[0]]["p95"])
    lines.append(
        f"  ➜ Endpoint terberat (P95): {heaviest[0]}  ({ep_data[heaviest[0]]['p95']:.0f} ms,"
        f" {ep_data[heaviest[0]]['p95'] / total_p95 * 100:.1f}% dari total flow)"
    )

lines.append("\n" + sep)
table = "\n".join(lines)
print(table)
(RESULT_DIR / "endpoint_breakdown_table.txt").write_text(table + "\n")
print(f"\nTabel disimpan: {RESULT_DIR / 'endpoint_breakdown_table.txt'}")


# ─── Chart ───────────────────────────────────────────────────────────────────
try:
    import matplotlib
    matplotlib.use("Agg")
    import matplotlib.pyplot as plt
except ImportError:
    print("\nmatplotlib tidak terinstall — chart dilewati.")
    sys.exit(0)

import numpy as np

EP_LABELS = [ep for ep, _ in ENDPOINTS]
TARGET_COLORS = {
    "Go echo":     "#0D47A1",
    "Go net/http": "#42A5F5",
    "Node.js":     "#4CAF50",
}


# Chart 1: grouped bar — P95 tiap endpoint, satu kelompok bar per target
def chart_p95():
    targets = list(results.keys())
    n = len(targets)
    x = np.arange(len(EP_LABELS))
    width = 0.8 / max(n, 1)

    fig, ax = plt.subplots(figsize=(12, 6))
    max_val = 0.0
    for i, target in enumerate(targets):
        vals = [results[target][ep]["p95"] for ep in EP_LABELS]
        max_val = max(max_val, *vals)
        offset = (i - (n - 1) / 2) * width
        bars = ax.bar(x + offset, vals, width,
                      label=target, color=TARGET_COLORS.get(target, "#888"),
                      alpha=0.9)
        for bar, v in zip(bars, vals):
            if v > 0:
                ax.text(bar.get_x() + bar.get_width() / 2, bar.get_height(),
                        f"{v:.0f}", ha="center", va="bottom", fontsize=9)

    ax.set_ylim(0, max_val * 1.18)
    ax.set_title("P95 Latency per Endpoint — Booking Flow", fontsize=14, fontweight="bold", pad=10)
    ax.set_ylabel("P95 Latency (ms)", fontsize=11)
    ax.set_xlabel("Endpoint", fontsize=11)
    ax.set_xticks(x)
    ax.set_xticklabels(EP_LABELS, fontsize=10)
    ax.grid(axis="y", alpha=0.3, linestyle="--")
    ax.legend(fontsize=10)
    ax.annotate("↓ better", xy=(0.98, 0.97), xycoords="axes fraction",
                ha="right", va="top", fontsize=10, color="gray", style="italic")

    plt.tight_layout()
    out = RESULT_DIR / "chart_endpoint_breakdown.png"
    plt.savefig(out, dpi=150, bbox_inches="tight")
    plt.close()
    print(f"Chart disimpan: {out}")


# Chart 2: stacked bar — kontribusi P95 tiap endpoint ke total flow per target
def chart_contribution():
    targets = list(results.keys())
    fig, ax = plt.subplots(figsize=(10, 6))

    ep_colors = {
        "check_availability": "#FFC107",  # kuning
        "create_reservation": "#E53935",  # merah (kemungkinan terberat)
        "pay_reservation":    "#FB8C00",  # oranye
        "cancel_reservation": "#1E88E5",  # biru
    }

    bottoms = np.zeros(len(targets))
    for ep in EP_LABELS:
        vals = np.array([results[t][ep]["p95"] for t in targets])
        bars = ax.bar(targets, vals, bottom=bottoms,
                      label=ep, color=ep_colors[ep], alpha=0.9)
        # Label nilai di tengah segmen
        for i, (b, v) in enumerate(zip(bottoms, vals)):
            if v > 30:  # skip label kalau segmen terlalu kecil
                ax.text(i, b + v / 2, f"{v:.0f}", ha="center", va="center",
                        fontsize=9, color="white", fontweight="bold")
        bottoms += vals

    # Total label di atas
    for i, total in enumerate(bottoms):
        ax.text(i, total + total * 0.02, f"Total: {total:.0f}ms",
                ha="center", va="bottom", fontsize=10, fontweight="bold")

    ax.set_ylim(0, max(bottoms) * 1.12 if len(bottoms) else 1)
    ax.set_title("Kontribusi tiap Endpoint ke Total Flow Latency (P95)",
                 fontsize=14, fontweight="bold", pad=10)
    ax.set_ylabel("Akumulasi P95 (ms)", fontsize=11)
    ax.set_xlabel("Target", fontsize=11)
    ax.grid(axis="y", alpha=0.3, linestyle="--")
    ax.legend(loc="upper right", fontsize=10, title="Endpoint")

    plt.tight_layout()
    out = RESULT_DIR / "chart_endpoint_contribution.png"
    plt.savefig(out, dpi=150, bbox_inches="tight")
    plt.close()
    print(f"Chart disimpan: {out}")


chart_p95()
chart_contribution()
print(f"\nSelesai! Semua file ada di: {RESULT_DIR}")
