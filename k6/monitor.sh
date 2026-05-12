#!/usr/bin/env bash
# monitor.sh
# Monitor CPU dan Memory proses Go selama pengujian k6
# Setara dengan monitor.ps1 untuk Linux (menggunakan /proc filesystem)
#
# Cara pakai:
#   # Terminal 1 — jalankan server Go
#   go run cmd/main.go
#
#   # Terminal 2 — jalankan monitor (SEBELUM k6 dimulai)
#   ./k6/monitor.sh --test-name "load_hotelbooking" --process-name "main"
#
#   # Terminal 3 — jalankan k6
#   k6 run k6/load_test.js
#
# Parameter:
#   --test-name     : nama file output CSV (tanpa ekstensi)
#   --process-name  : nama proses yang dimonitor (default: "main" untuk Go)
#   --pid           : monitor berdasarkan PID spesifik (opsional)
#   --interval      : interval sampling dalam detik (default: 1)
#   --duration      : durasi monitoring dalam detik (default: 600 = 10 menit)

TEST_NAME="monitor"
PROCESS_NAME="main"
PROCESS_ID=0
INTERVAL_SEC=1
DURATION_SEC=600

while [[ $# -gt 0 ]]; do
    case $1 in
        --test-name|-t)    TEST_NAME="$2";    shift 2 ;;
        --process-name|-n) PROCESS_NAME="$2"; shift 2 ;;
        --pid|-p)          PROCESS_ID="$2";   shift 2 ;;
        --interval|-i)     INTERVAL_SEC="$2"; shift 2 ;;
        --duration|-d)     DURATION_SEC="$2"; shift 2 ;;
        *) echo "Unknown option: $1"; exit 1 ;;
    esac
done

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
OUTPUT_DIR="$SCRIPT_DIR/results"
mkdir -p "$OUTPUT_DIR"

OUTPUT_FILE="$OUTPUT_DIR/${TEST_NAME}_monitor.csv"
SUMMARY_FILE="$OUTPUT_DIR/${TEST_NAME}_cpu_mem_summary.txt"
CLK_TCK=$(getconf CLK_TCK)
NUM_CORES=$(nproc)

echo "Timestamp,CPU_Percent,Memory_MB,Working_Set_MB,ProcessName" > "$OUTPUT_FILE"

find_pid() {
    if [[ $PROCESS_ID -gt 0 ]]; then
        echo "$PROCESS_ID"
    else
        # Ambil PID dengan memory terbesar jika ada multiple proses
        pgrep -x "$PROCESS_NAME" 2>/dev/null | while read -r pid; do
            vmrss=$(awk '/^VmRSS:/{print $2}' "/proc/$pid/status" 2>/dev/null || echo 0)
            echo "$vmrss $pid"
        done | sort -rn | awk '{print $2; exit}'
    fi
}

# Ambil total CPU jiffies (utime+stime) dari /proc/$pid/stat
get_cpu_jiffies() {
    local pid=$1
    local stat_file="/proc/$pid/stat"
    [[ -f "$stat_file" ]] || { echo 0; return 1; }
    local content after_comm fields
    content=$(cat "$stat_file")
    after_comm="${content#*) }"
    fields=($after_comm)
    # Setelah strip "pid (comm) ": index 11=utime, 12=stime (0-indexed)
    echo $(( ${fields[11]} + ${fields[12]} ))
}

# Ambil memory dalam MB dari /proc/$pid/status
# Output: "vmsize_mb vmrss_mb"
get_mem_mb() {
    local pid=$1
    local status_file="/proc/$pid/status"
    [[ -f "$status_file" ]] || { echo "0 0"; return 1; }
    awk '/^VmSize:/{vmsize=$2} /^VmRSS:/{vmrss=$2}
         END{printf "%.2f %.2f\n", vmsize/1024, vmrss/1024}' "$status_file"
}

print_summary() {
    [[ ! -f "$OUTPUT_FILE" ]] && return
    local line_count
    line_count=$(awk 'NR>1{count++} END{print count+0}' "$OUTPUT_FILE")
    [[ $line_count -eq 0 ]] && return

    echo ""
    echo "=================================================="
    echo "  HASIL MONITORING: $TEST_NAME"
    echo "=================================================="

    awk -F',' 'NR>1 {
        cpu+=$2; if($2>cpu_max) cpu_max=$2;
        mem+=$3; if($3>mem_max) mem_max=$3;
        n++
    }
    END{
        printf "  CPU Rata-rata : %.2f %%\n",  cpu/n
        printf "  CPU Maksimum  : %.2f %%\n",  cpu_max
        printf "  Mem Rata-rata : %.2f MB\n",  mem/n
        printf "  Mem Maksimum  : %.2f MB\n",  mem_max
        printf "  Total Sampel  : %d\n",        n
    }' "$OUTPUT_FILE"

    awk -F',' -v test="$TEST_NAME" -v proc="$PROCESS_NAME" 'NR>1 {
        cpu+=$2; if($2>cpu_max) cpu_max=$2;
        mem+=$3; if($3>mem_max) mem_max=$3;
        n++
    }
    END{
        printf "=== CPU & MEMORY SUMMARY: %s ===\n", test
        printf "Proses      : %s\n",   proc
        printf "Sampel      : %d\n",   n
        printf "CPU Avg     : %.2f %%\n", cpu/n
        printf "CPU Max     : %.2f %%\n", cpu_max
        printf "Mem Avg     : %.2f MB\n", mem/n
        printf "Mem Max     : %.2f MB\n", mem_max
    }' "$OUTPUT_FILE" > "$SUMMARY_FILE"

    echo ""
    echo "  Ringkasan disimpan ke: $SUMMARY_FILE"
    echo "  CSV data    : $OUTPUT_FILE"
    echo "=================================================="
}

cleanup() {
    echo ""
    echo "Monitor dihentikan (Ctrl+C)"
    print_summary
    exit 0
}
trap cleanup INT TERM

echo ""
echo "=================================================="
echo "  CPU & Memory Monitor"
echo "  Target Process : $PROCESS_NAME"
echo "  Output         : $OUTPUT_FILE"
echo "  Interval       : ${INTERVAL_SEC}s | Durasi: ${DURATION_SEC}s"
echo "  Ctrl+C untuk berhenti lebih awal"
echo "=================================================="
echo ""

declare -A prev_jiffies
sample_count=0
end_time=$(( $(date +%s) + DURATION_SEC ))

while [[ $(date +%s) -lt $end_time ]]; do
    timestamp=$(date '+%Y-%m-%d %H:%M:%S')
    pid=$(find_pid)

    if [[ -z "$pid" ]]; then
        label="'$PROCESS_NAME'"
        [[ $PROCESS_ID -gt 0 ]] && label="PID $PROCESS_ID"
        echo "[$timestamp] Proses $label tidak ditemukan, menunggu..."
        sleep "$INTERVAL_SEC"
        continue
    fi

    jiffies=$(get_cpu_jiffies "$pid") || { sleep "$INTERVAL_SEC"; continue; }

    cpu_percent=0
    if [[ -n "${prev_jiffies[$pid]+_}" ]]; then
        delta=$(( jiffies - prev_jiffies[$pid] ))
        cpu_percent=$(awk -v d="$delta" -v clk="$CLK_TCK" -v iv="$INTERVAL_SEC" -v c="$NUM_CORES" \
            'BEGIN{ v=(d/clk/iv/c)*100; if(v<0)v=0; if(v>100)v=100; printf "%.2f", v }')
    fi
    prev_jiffies[$pid]=$jiffies

    read -r vmsize_mb vmrss_mb < <(get_mem_mb "$pid")
    proc_name=$(cat "/proc/$pid/comm" 2>/dev/null || echo "$PROCESS_NAME")

    echo "$timestamp,$cpu_percent,$vmsize_mb,$vmrss_mb,$proc_name" >> "$OUTPUT_FILE"
    sample_count=$(( sample_count + 1 ))

    if (( sample_count % 5 == 0 )); then
        echo "[$timestamp] CPU: ${cpu_percent}% | Mem: ${vmsize_mb} MB | WS: ${vmrss_mb} MB"
    else
        echo "[$timestamp] CPU: ${cpu_percent}% | Mem: ${vmsize_mb} MB"
    fi

    sleep "$INTERVAL_SEC"
done

print_summary
