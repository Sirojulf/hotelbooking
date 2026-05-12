# run_compare.ps1
# Jalankan kedua k6 test secara berurutan, lalu tampilkan perbandingan
#
# Cara pakai:
#   cd d:\dev\hotelbooking
#   .\k6\run_compare.ps1

$ErrorActionPreference = 'Stop'

# Buat folder hasil kalau belum ada
New-Item -ItemType Directory -Force -Path "k6\results" | Out-Null

Write-Host ""
Write-Host "══════════════════════════════════════════════════" -ForegroundColor Cyan
Write-Host "  [1/2] Menjalankan test hotelbooking (REST API)" -ForegroundColor Cyan
Write-Host "══════════════════════════════════════════════════" -ForegroundColor Cyan
k6 run k6/hotelbooking_booking.js
if ($LASTEXITCODE -ne 0) {
    Write-Host "  hotelbooking test gagal (exit $LASTEXITCODE)" -ForegroundColor Red
}

Write-Host ""
Write-Host "══════════════════════════════════════════════════" -ForegroundColor Cyan
Write-Host "  [2/2] Menjalankan test RoomMasterb (Browser)" -ForegroundColor Cyan
Write-Host "══════════════════════════════════════════════════" -ForegroundColor Cyan
k6 run k6/roommaster_booking.js
if ($LASTEXITCODE -ne 0) {
    Write-Host "  RoomMasterb test gagal (exit $LASTEXITCODE)" -ForegroundColor Red
}

Write-Host ""
Write-Host "══════════════════════════════════════════════════" -ForegroundColor Green
Write-Host "  Membandingkan hasil..." -ForegroundColor Green
Write-Host "══════════════════════════════════════════════════" -ForegroundColor Green
node k6/compare_results.js
