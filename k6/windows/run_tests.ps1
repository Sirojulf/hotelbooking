# run_tests.ps1
# Jalankan semua 3 jenis test (Load, Spike, Stress) untuk satu target API
# Setiap test: monitor CPU/Mem berjalan di background, k6 di foreground
#
# Cara pakai:
#   # Test hotelbooking (Go):
#   .\k6\run_tests.ps1 -Target hotelbooking
#
#   # Test RoomMasterb:
#   .\k6\run_tests.ps1 -Target roommaster
#
# Pastikan:
#   1. Server target sudah berjalan
#   2. config.js sudah diisi dengan UUID yang benar
#   3. k6 sudah terinstall (winget install k6)

param(
    [ValidateSet("hotelbooking","roommaster")]
    [string]$Target = "hotelbooking",

    [string]$ProcessName = "main",   # nama proses server yang dimonitor
    [switch]$SkipMonitor             # lewati monitoring CPU/Mem (jika tidak perlu)
)

$ErrorActionPreference = "Continue"
$ResultDir = Join-Path $PSScriptRoot "results"
New-Item -ItemType Directory -Force -Path $ResultDir | Out-Null

function Write-Section($msg) {
    Write-Host ""
    Write-Host ("=" * 52) -ForegroundColor Cyan
    Write-Host "  $msg" -ForegroundColor Cyan
    Write-Host ("=" * 52) -ForegroundColor Cyan
}

function Run-K6Test($testFile, $testName, $outputJson) {
    Write-Section "$testName — Target: $Target"

    # Mulai monitor CPU/Mem di background
    $monitorJob = $null
    if (-not $SkipMonitor) {
        $monitorScript = Join-Path $PSScriptRoot "monitor.ps1"
        $monitorJob = Start-Job -ScriptBlock {
            param($script, $name, $proc)
            & pwsh -NonInteractive -File $script -TestName $name -ProcessName $proc -DurationSec 900
        } -ArgumentList $monitorScript, "${Target}_$testName", $ProcessName

        Write-Host "  Monitor CPU/Mem dimulai (job ID: $($monitorJob.Id))" -ForegroundColor Yellow
        Start-Sleep -Seconds 3  # beri waktu monitor untuk mulai
    }

    # Jalankan k6
    Write-Host "  Menjalankan: k6 run --out json=$outputJson $testFile" -ForegroundColor White
    $k6Start = Get-Date

    # Edit file untuk mengarahkan ke target yang benar jika TARGET=roommaster
    # (k6 tidak support -e untuk switch import, jadi kita pakai env var dan cek di script)
    & k6 run --out "json=$outputJson" $testFile
    $k6ExitCode = $LASTEXITCODE

    $k6Duration = [math]::Round(((Get-Date) - $k6Start).TotalMinutes, 1)
    Write-Host "  k6 selesai dalam $k6Duration menit (exit: $k6ExitCode)" -ForegroundColor $(if ($k6ExitCode -eq 0) { "Green" } else { "Red" })

    # Hentikan monitor
    if ($monitorJob) {
        Stop-Job -Job $monitorJob -ErrorAction SilentlyContinue
        Remove-Job -Job $monitorJob -ErrorAction SilentlyContinue
        Write-Host "  Monitor dihentikan" -ForegroundColor Yellow
    }

    # Jeda antar test (sistem stabilisasi)
    Write-Host "  Menunggu 20 detik untuk stabilisasi sistem..." -ForegroundColor Gray
    Start-Sleep -Seconds 20

    return $k6ExitCode
}

# ─── Cek k6 terinstall ───────────────────────────────────────────────────────
if (-not (Get-Command k6 -ErrorAction SilentlyContinue)) {
    Write-Host "ERROR: k6 tidak ditemukan. Install dengan:" -ForegroundColor Red
    Write-Host "  winget install k6" -ForegroundColor Yellow
    exit 1
}

# ─── Cek server target berjalan ──────────────────────────────────────────────
Write-Section "Persiapan Test — Target: $Target"

$healthURL = if ($Target -eq "hotelbooking") { "http://localhost:8080/health" } else { "http://localhost:3000/api/health" }
try {
    $health = Invoke-RestMethod -Uri $healthURL -TimeoutSec 5 -ErrorAction Stop
    Write-Host "  Server berjalan: $healthURL" -ForegroundColor Green
} catch {
    Write-Host "  PERINGATAN: Health check gagal ($healthURL)" -ForegroundColor Red
    Write-Host "  Pastikan server $Target sudah berjalan sebelum melanjutkan!" -ForegroundColor Yellow
    $confirm = Read-Host "  Lanjutkan anyway? (y/N)"
    if ($confirm -ne 'y' -and $confirm -ne 'Y') { exit 1 }
}

# ─── Jalankan 3 test ─────────────────────────────────────────────────────────
$k6Dir      = $PSScriptRoot
$loadScript = Join-Path $k6Dir "load_test.js"
$spikeScript = Join-Path $k6Dir "spike_test.js"
$stressScript = Join-Path $k6Dir "stress_test.js"

$results = @{}

# 1. Load Test
$results["load"] = Run-K6Test `
    -testFile  $loadScript `
    -testName  "load" `
    -outputJson (Join-Path $ResultDir "load_${Target}.json")

# 2. Spike Test
$results["spike"] = Run-K6Test `
    -testFile  $spikeScript `
    -testName  "spike" `
    -outputJson (Join-Path $ResultDir "spike_${Target}.json")

# 3. Stress Test
$results["stress"] = Run-K6Test `
    -testFile  $stressScript `
    -testName  "stress" `
    -outputJson (Join-Path $ResultDir "stress_${Target}.json")

# ─── Ringkasan Akhir ─────────────────────────────────────────────────────────
Write-Section "SELESAI — Semua Test untuk: $Target"
foreach ($test in $results.Keys) {
    $status = if ($results[$test] -eq 0) { "LULUS" } else { "GAGAL THRESHOLD" }
    $color  = if ($results[$test] -eq 0) { "Green" } else { "Yellow" }
    Write-Host "  $($test.ToUpper().PadRight(8)): $status (exit $($results[$test]))" -ForegroundColor $color
}
Write-Host ""
Write-Host "  Hasil disimpan di: $ResultDir" -ForegroundColor Cyan
Write-Host "  File summary JSON : load_${Target}.json, spike_${Target}.json, stress_${Target}.json"

# Tampilkan ringkasan CPU/Mem jika ada
if (-not $SkipMonitor) {
    Write-Host ""
    Write-Host "  Ringkasan CPU & Memory:" -ForegroundColor Cyan
    Get-ChildItem -Path $ResultDir -Filter "*${Target}*cpu_mem_summary.txt" | ForEach-Object {
        Write-Host ""
        Get-Content $_.FullName | ForEach-Object { Write-Host "  $_" }
    }
}
