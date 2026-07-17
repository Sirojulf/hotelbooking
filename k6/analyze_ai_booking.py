"""Generate an IEEE-ready AI booking cost projection.

The estimate uses the average OpenAI iterations reported by the k6 AI booking
test and explicit token/pricing assumptions. Only the cost figure is produced.

Run after:
    k6 run ai_booking_test.js

Usage:
    python3 analyze_ai_booking.py

Output:
    results/chart_ai_cost.{pdf,png}
"""

import json
import sys
from pathlib import Path

RESULT_DIR = Path(__file__).parent / "results"

# Cost assumptions. Update these constants when the model or pricing changes.
MODEL_NAME = "gpt-4o-mini"
PRICE_INPUT_PER_1M = 0.15
PRICE_OUTPUT_PER_1M = 0.60
TOKENS_IN_PER_CALL = 1500
TOKENS_OUT_PER_CALL = 200

BOOKING_VOLUMES = [100, 1_000, 10_000, 100_000, 1_000_000]
VOLUME_LABELS = ["100", "1K", "10K", "100K", "1M"]

IEEE_COLUMN_WIDTH = 3.50
IEEE_FIGURE_HEIGHT = 2.65
PNG_DPI = 600
LINE_COLOR = "#FF9800"


def load_ai_iterations():
    """Return the average model calls per booking from the k6 summary."""
    filepath = RESULT_DIR / "ai_booking.json"
    if not filepath.exists():
        print(f"ERROR: input file not found: {filepath}")
        print("Run this first: k6 run ai_booking_test.js")
        sys.exit(1)

    try:
        data = json.loads(filepath.read_text(encoding="utf-8"))
    except json.JSONDecodeError as exc:
        print(f"ERROR: invalid JSON in {filepath}: {exc}")
        sys.exit(1)

    iterations = (
        data.get("metrics", {})
        .get("ai_openai_iterations", {})
        .get("values", {})
        .get("avg", 0)
    )
    if not isinstance(iterations, (int, float)) or iterations <= 0:
        print("ERROR: ai_openai_iterations.avg is missing or invalid.")
        sys.exit(1)
    return float(iterations)


def calculate_costs(iterations_per_booking):
    """Calculate the estimated cost at each configured booking volume."""
    cost_per_call = (
        TOKENS_IN_PER_CALL / 1_000_000 * PRICE_INPUT_PER_1M
        + TOKENS_OUT_PER_CALL / 1_000_000 * PRICE_OUTPUT_PER_1M
    )
    cost_per_booking = cost_per_call * iterations_per_booking
    return cost_per_call, cost_per_booking, [
        cost_per_booking * volume for volume in BOOKING_VOLUMES
    ]


def format_currency(value):
    """Format chart labels using precision appropriate to the cost."""
    if value < 100:
        return f"${value:.2f}"
    return f"${value:,.0f}"


def create_cost_figure(iterations_per_booking, costs):
    """Create a single-column IEEE cost projection figure."""
    try:
        import matplotlib

        matplotlib.use("Agg")
        import matplotlib.pyplot as plt
    except ImportError:
        print("matplotlib is not installed; the cost figure was not generated.")
        sys.exit(1)

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
            "axes.linewidth": 0.6,
            "lines.linewidth": 1.1,
            "figure.facecolor": "white",
            "axes.facecolor": "white",
            "savefig.facecolor": "white",
            "pdf.fonttype": 42,
            "ps.fonttype": 42,
            "mathtext.fontset": "stix",
        }
    )

    x_positions = list(range(len(VOLUME_LABELS)))
    fig, ax = plt.subplots(figsize=(IEEE_COLUMN_WIDTH, IEEE_FIGURE_HEIGHT))
    ax.plot(
        x_positions,
        costs,
        color=LINE_COLOR,
        marker="o",
        markersize=4.5,
        markeredgecolor="black",
        markeredgewidth=0.5,
        zorder=3,
    )

    for x_position, cost in zip(x_positions, costs):
        ax.annotate(
            format_currency(cost),
            (x_position, cost),
            xytext=(0, 5),
            textcoords="offset points",
            ha="center",
            va="bottom",
            fontsize=6.5,
        )

    ax.set_yscale("log")
    ax.set_title("Estimated AI Booking Cost", fontweight="normal", pad=4)
    ax.set_xlabel("Number of Bookings")
    ax.set_ylabel("Estimated Cost (USD, log scale)")
    ax.set_xticks(x_positions)
    ax.set_xticklabels(VOLUME_LABELS)
    ax.tick_params(axis="both", direction="out", length=2.5, width=0.6)
    ax.grid(axis="y", which="both", color="#BFBFBF", linewidth=0.45, linestyle=":")
    ax.spines["top"].set_visible(False)
    ax.spines["right"].set_visible(False)
    ax.text(
        0.02,
        0.97,
        f"{MODEL_NAME}; {iterations_per_booking:.1f} calls/booking",
        transform=ax.transAxes,
        ha="left",
        va="top",
        fontsize=6.5,
        color="#444444",
    )

    fig.subplots_adjust(left=0.18, right=0.98, bottom=0.20, top=0.86)
    output_paths = []
    for extension, dpi in (("pdf", None), ("png", PNG_DPI)):
        outpath = RESULT_DIR / f"chart_ai_cost.{extension}"
        fig.savefig(outpath, dpi=dpi, metadata={"Creator": "analyze_ai_booking.py"})
        output_paths.append(outpath)
    plt.close(fig)
    print(f"Figures saved: {output_paths[0]} and {output_paths[1]}")


def main():
    iterations_per_booking = load_ai_iterations()
    cost_per_call, cost_per_booking, costs = calculate_costs(iterations_per_booking)

    print("AI booking cost assumptions:")
    print(f"  Model: {MODEL_NAME}")
    print(f"  Average model calls per booking: {iterations_per_booking:.2f}")
    print(f"  Estimated cost per model call: ${cost_per_call:.6f}")
    print(f"  Estimated cost per booking: ${cost_per_booking:.6f}")
    create_cost_figure(iterations_per_booking, costs)


if __name__ == "__main__":
    main()
