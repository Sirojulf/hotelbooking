#!/usr/bin/env bash
# run_tests.sh
# Jalankan semua 3 jenis test (Load, Spike, Stress) untuk satu target API
# Setiap test: monitor CPU/Mem berjalan di background, k6 di foreground
#
# Cara pakai:
#   # Test hotelbooking (Go) — dari root project:
#   ./k6/run_tests.sh hotelbooking
#
#   # Test RoomMaster:
#   ./k6/run_tests.sh roommaster
#
#   # Lewati monitoring:
#   ./k6/run_tests.sh hotelbooking main skip
#
# Pastikan:
#   1. Server target sudah berjalan
#   2. config.js sudah diisi dengan UUID yang benar
#   3. k6 sudah terinstall (lihat: https://grafana.com/docs/k6/latest/set-up/install-k6/)

TARGET="${1:-hotelbooking}"
PROCESS_NAME="${2:-main}"
SKIP_MONITOR="${3:-}"   # pass "skip" untuk lewati monitoring

if [[ "$TARGET" != "hotelbooking" && "$TARGET" != "roommaster" ]]; then
    echo "Usage: $0 [hotelbooking|roommaster] [process-name] [skip]"
    exit 1
fi

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
RESULT_DIR="$SCRIPT_DIR/results"
mkdir -p "$RESULT_DIR"

# ─── Warna output ────────────────────────────────────────────────────────────
cyan()   { echo -e "\033[36m$*\033[0m"; }
green()  { echo -e "\033[32m$*\033[0m"; }
yellow() { echo -e "\033[33m$*\033[0m"; }
red()    { echo -e "\033[31m$*\033[0m"; }
gray()   { echo -e "\033[90m$*\033[0m"; }

section() {
    echo ""
    cyan "===================================================="
    cyan "  $1"
    cyan "===================================================="
}

run_k6_test() {
    local test_file="$1"
    local test_name="$2"
    local output_json="$3"

    section "$test_name — Target: $TARGET"

    # Mulai monitor CPU/Mem di background
    local monitor_pid=""
    if [[ -z "$SKIP_MONITOR" ]]; then
        local monitor_script="$SCRIPT_DIR/monitor.sh"
        if [[ -x "$monitor_script" ]]; then
            "$monitor_script" \
                --test-name "${test_name}_${TARGET}" \
                --process-name "$PROCESS_NAME" \
                --port "$SERVER_PORT" \
                --duration 900 &
            monitor_pid=$!
            yellow "  Monitor CPU/Mem dimulai (PID: $monitor_pid)"
            sleep 3
        else
            yellow "  monitor.sh tidak ditemukan atau tidak executable, skip monitoring"
        fi
    fi

    # Jalankan k6
    echo "  Menjalankan: k6 run --out json=$output_json $test_file"
    local k6_start k6_end duration_min k6_exit
    k6_start=$(date +%s)

    k6 run -e TARGET="$TARGET" --out "json=$output_json" "$test_file"
    k6_exit=$?

    k6_end=$(date +%s)
    duration_min=$(awk -v s=$(( k6_end - k6_start )) 'BEGIN{printf "%.1f", s/60}')

    if [[ $k6_exit -eq 0 ]]; then
        green "  k6 selesai dalam $duration_min menit (exit: $k6_exit)"
    else
        red "  k6 selesai dalam $duration_min menit (exit: $k6_exit)"
    fi

    # Hentikan monitor
    if [[ -n "$monitor_pid" ]]; then
        kill "$monitor_pid" 2>/dev/null || true
        wait "$monitor_pid" 2>/dev/null || true
        yellow "  Monitor dihentikan"
    fi

    gray "  Menunggu 20 detik untuk stabilisasi sistem..."
    sleep 20

    return $k6_exit
}

# ─── Cek k6 terinstall ───────────────────────────────────────────────────────
if ! command -v k6 &>/dev/null; then
    red "ERROR: k6 tidak ditemukan. Install dengan:"
    yellow "  # Fedora/RHEL:"
    yellow "  sudo dnf install https://dl.k6.io/rpm/repo.rpm && sudo dnf install k6"
    yellow "  # Ubuntu/Debian:"
    yellow "  sudo gpg --no-default-keyring --keyring /usr/share/keyrings/k6-archive-keyring.gpg \\"
    yellow "       --keyserver hkp://keyserver.ubuntu.com:80 --recv-keys C5AD17C747E3415A3642D57D77C6C491D6AC1D69"
    yellow "  echo 'deb [signed-by=/usr/share/keyrings/k6-archive-keyring.gpg] https://dl.k6.io/deb stable main' \\"
    yellow "       | sudo tee /etc/apt/sources.list.d/k6.list"
    yellow "  sudo apt-get update && sudo apt-get install k6"
    exit 1
fi

# ─── Cek server target berjalan ──────────────────────────────────────────────
section "Persiapan Test — Target: $TARGET"

if [[ "$TARGET" == "hotelbooking" ]]; then
    health_url="http://localhost:8080/health"
    SERVER_PORT=8080
else
    health_url="http://localhost:3000/api/health"
    SERVER_PORT=3000
fi

if curl -sf --max-time 5 "$health_url" &>/dev/null; then
    green "  Server berjalan: $health_url"
else
    red "  PERINGATAN: Health check gagal ($health_url)"
    yellow "  Pastikan server $TARGET sudah berjalan sebelum melanjutkan!"
    read -rp "  Lanjutkan anyway? (y/N): " confirm
    [[ "$confirm" =~ ^[yY]$ ]] || exit 1
fi

# ─── Jalankan 3 test ─────────────────────────────────────────────────────────
declare -A results

# 1. Load Test
run_k6_test \
    "$SCRIPT_DIR/load_test.js" \
    "load" \
    "$RESULT_DIR/load_${TARGET}.json"
results["load"]=$?

# 2. Spike Test
run_k6_test \
    "$SCRIPT_DIR/spike_test.js" \
    "spike" \
    "$RESULT_DIR/spike_${TARGET}.json"
results["spike"]=$?

# 3. Stress Test
run_k6_test \
    "$SCRIPT_DIR/stress_test.js" \
    "stress" \
    "$RESULT_DIR/stress_${TARGET}.json"
results["stress"]=$?

# ─── Ringkasan Akhir ─────────────────────────────────────────────────────────
section "SELESAI — Semua Test untuk: $TARGET"
for test in load spike stress; do
    exit_code=${results[$test]}
    label=$(printf '%-8s' "${test^^}")
    if [[ $exit_code -eq 0 ]]; then
        green "  $label: LULUS (exit $exit_code)"
    else
        yellow "  $label: GAGAL THRESHOLD (exit $exit_code)"
    fi
done

echo ""
cyan "  Hasil disimpan di: $RESULT_DIR"
echo "  File JSON: load_${TARGET}.json, spike_${TARGET}.json, stress_${TARGET}.json"

# Tampilkan ringkasan CPU/Mem jika monitoring aktif
if [[ -z "$SKIP_MONITOR" ]]; then
    echo ""
    cyan "  Ringkasan CPU & Memory:"
    for f in "$RESULT_DIR"/*"${TARGET}"*cpu_mem_summary.txt; do
        [[ -f "$f" ]] || continue
        echo ""
        while IFS= read -r line; do echo "  $line"; done < "$f"
    done
fi
