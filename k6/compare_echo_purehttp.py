"""
Bandingkan hasil k6 antara versi Echo vs pure net/http.
Jalankan SETELAH backup ke k6/results/echo/ dan k6/results/purehttp/

Cara pakai:
    python3 k6/compare_echo_purehttp.py
"""
import json
import math
from pathlib import Path

RESULT_DIR = Path(__file__).parent / "results"
TESTS = ["load", "spike", "stress"]


def parse_k6_result(filepath):
    """Parse hasil k6. Coba format summary JSON dulu (dari handleSummary),
    fallback ke NDJSON (dari --out json=)."""
    if not filepath.exists():
        return None
    raw = filepath.read_text().strip()
    if not raw:
        return None

    # Format 1: summary JSON (single object dengan key "metrics")
    try:
        data = json.loads(raw)
        if isinstance(data, dict) and "metrics" in data:
            return _from_summary(data["metrics"])
    except json.JSONDecodeError:
        pass

    # Format 2: NDJSON (satu objek per baris)
    return _from_ndjson(raw)


def _from_summary(metrics):
    """Ekstrak metrik dari format summary JSON k6."""
    def v(key, stat):
        return metrics.get(key, {}).get("values", {}).get(stat, 0)

    return {
        "total_reqs": v("http_reqs", "count"),
        "error_rate": v("http_req_failed", "rate") * 100,
        "success_rate": v("booking_success_rate", "rate") * 100,
        "p50": v("http_req_duration", "med"),
        "p95": v("http_req_duration", "p(95)"),
        "p99": v("http_req_duration", "p(99)"),
        "flow_p95": v("booking_flow_duration_ms", "p(95)"),
    }


def _from_ndjson(raw):
    """Ekstrak metrik dari format NDJSON k6 (--out json=)."""
    durations, flow_durations = [], []
    failed, total_reqs = 0, 0
    succ_rates = []
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
        "error_rate": (failed / total_reqs * 100) if total_reqs else 0,
        "success_rate": (sum(succ_rates) / len(succ_rates) * 100) if succ_rates else 0,
        "p50": pct(durations, 50),
        "p95": pct(durations, 95),
        "p99": pct(durations, 99),
        "flow_p95": pct(flow_durations, 95),
    }


def read_monitor(folder, test):
    f = RESULT_DIR / folder / f"{test}_hotelbooking_cpu_mem_summary.txt"
    res = {"cpu_avg": "N/A", "cpu_max": "N/A", "mem_avg": "N/A", "mem_max": "N/A"}
    if not f.exists():
        return res
    for line in f.read_text().splitlines():
        if "CPU Avg" in line:
            res["cpu_avg"] = line.split(":")[-1].strip()
        elif "CPU Max" in line:
            res["cpu_max"] = line.split(":")[-1].strip()
        elif "Mem Avg" in line:
            res["mem_avg"] = line.split(":")[-1].strip()
        elif "Mem Max" in line:
            res["mem_max"] = line.split(":")[-1].strip()
    return res


print("\n" + "=" * 78)
print("  PERBANDINGAN: Echo  vs  pure net/http  (logic identik, beda framework)")
print("=" * 78)

for test in TESTS:
    echo = parse_k6_result(RESULT_DIR / "echo" / f"{test}_hotelbooking.json")
    pure = parse_k6_result(RESULT_DIR / "purehttp" / f"{test}_hotelbooking.json")
    if not echo or not pure:
        print(f"\n  [{test.upper()}] data tidak lengkap — pastикan sudah backup ke echo/ & purehttp/")
        continue
    em = read_monitor("echo", test)
    pm = read_monitor("purehttp", test)

    print(f"\n  [ {test.upper()} TESTING ]")
    print("  " + "-" * 76)
    print(f"  {'Metrik':<26}{'Echo':>16}{'pure net/http':>18}{'Selisih':>14}")
    print("  " + "-" * 76)

    def row(label, e, p, unit="", lower_better=True):
        try:
            ev, pv = float(e), float(p)
            if ev == 0:
                diff = "-"
            else:
                d = (pv - ev) / ev * 100
                better = (d < 0) if lower_better else (d > 0)
                mark = "✓" if better else "✗"
                diff = f"{d:+.1f}% {mark}"
        except (ValueError, TypeError):
            diff = "-"
        print(f"  {label:<26}{str(e)+unit:>16}{str(p)+unit:>18}{diff:>14}")

    row("Total HTTP Requests", echo["total_reqs"], pure["total_reqs"], "", lower_better=False)
    row("Error Rate", f"{echo['error_rate']:.2f}", f"{pure['error_rate']:.2f}", "%")
    row("Booking Success Rate", f"{echo['success_rate']:.1f}", f"{pure['success_rate']:.1f}", "%", lower_better=False)
    row("HTTP P50", f"{echo['p50']:.0f}", f"{pure['p50']:.0f}", "ms")
    row("HTTP P95", f"{echo['p95']:.0f}", f"{pure['p95']:.0f}", "ms")
    row("HTTP P99", f"{echo['p99']:.0f}", f"{pure['p99']:.0f}", "ms")
    row("Flow P95", f"{echo['flow_p95']:.0f}", f"{pure['flow_p95']:.0f}", "ms")
    row("CPU Avg", em["cpu_avg"].replace(" %", ""), pm["cpu_avg"].replace(" %", ""), " %")
    row("CPU Max", em["cpu_max"].replace(" %", ""), pm["cpu_max"].replace(" %", ""), " %")
    row("Mem Avg", em["mem_avg"].replace(" MB", ""), pm["mem_avg"].replace(" MB", ""), " MB")
    row("Mem Max", em["mem_max"].replace(" MB", ""), pm["mem_max"].replace(" MB", ""), " MB")

print("\n" + "=" * 78)
print("  ✓ = pure net/http lebih baik    ✗ = Echo lebih baik")
print("=" * 78 + "\n")
