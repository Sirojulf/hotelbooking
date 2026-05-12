#!/usr/bin/env bash
# run_compare.sh
# Jalankan kedua k6 test secara berurutan, lalu tampilkan perbandingan
#
# Cara pakai:
#   cd /path/to/hotelbooking
#   ./k6/run_compare.sh

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

cyan()  { echo -e "\033[36m$*\033[0m"; }
green() { echo -e "\033[32m$*\033[0m"; }
red()   { echo -e "\033[31m$*\033[0m"; }

mkdir -p "$SCRIPT_DIR/results"

echo ""
cyan "════════════════════════════════════════════════════"
cyan "  [1/2] Menjalankan test hotelbooking (REST API)"
cyan "════════════════════════════════════════════════════"
k6 run "$SCRIPT_DIR/hotelbooking_booking.js"
hb_exit=$?
[[ $hb_exit -ne 0 ]] && red "  hotelbooking test gagal (exit $hb_exit)"

echo ""
cyan "════════════════════════════════════════════════════"
cyan "  [2/2] Menjalankan test RoomMaster (Browser)"
cyan "════════════════════════════════════════════════════"
k6 run "$SCRIPT_DIR/roommaster_booking.js"
rm_exit=$?
[[ $rm_exit -ne 0 ]] && red "  RoomMaster test gagal (exit $rm_exit)"

echo ""
green "════════════════════════════════════════════════════"
green "  Membandingkan hasil..."
green "════════════════════════════════════════════════════"
node "$SCRIPT_DIR/compare_results.js"
