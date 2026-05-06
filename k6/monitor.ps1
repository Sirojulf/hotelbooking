# monitor.ps1
# Monitor CPU dan Memory penggunaan Go server selama pengujian k6
# Setara dengan Glances yang digunakan di referensi PDF (untuk Windows)
#
# Cara pakai:
#   # Terminal 1 — jalankan server Go
#   go run cmd/main.go
#
#   # Terminal 2 — jalankan monitor (SEBELUM k6 dimulai)
#   .\k6\monitor.ps1 -TestName "load_hotelbooking" -ProcessName "main"
#
#   # Terminal 3 — jalankan k6
#   k6 run k6/load_test.js
#
# Parameter:
#   -TestName     : nama file output CSV (tanpa ekstensi)
#   -ProcessName  : nama proses yang dimonitor (default: "main" untuk Go)
#   -IntervalSec  : interval sampling dalam detik (default: 1)
#   -DurationSec  : durasi monitoring dalam detik (default: 600 = 10 menit)

param(
    [string]$TestName    = "monitor",
    [string]$ProcessName = "main",
    [int]$IntervalSec    = 1,
    [int]$DurationSec    = 600
)

$OutputDir = Join-Path $PSScriptRoot "results"
New-Item -ItemType Directory -Force -Path $OutputDir | Out-Null

$OutputFile = Join-Path $OutputDir "$TestName`_monitor.csv"

# Header CSV
"Timestamp,CPU_Percent,Memory_MB,Working_Set_MB,ProcessName" | Out-File -FilePath $OutputFile -Encoding UTF8

$startTime  = Get-Date
$endTime    = $startTime.AddSeconds($DurationSec)
$sampleCount = 0

Write-Host ""
Write-Host "==================================================" -ForegroundColor Cyan
Write-Host "  CPU & Memory Monitor" -ForegroundColor Cyan
Write-Host "  Target Process : $ProcessName" -ForegroundColor Cyan
Write-Host "  Output         : $OutputFile" -ForegroundColor Cyan
Write-Host "  Interval       : ${IntervalSec}s | Durasi: ${DurationSec}s" -ForegroundColor Cyan
Write-Host "  Ctrl+C untuk berhenti lebih awal" -ForegroundColor Yellow
Write-Host "==================================================" -ForegroundColor Cyan
Write-Host ""

# Untuk kalkulasi CPU — simpan nilai CPU counter sebelumnya
$prevCpuTime = @{}

try {
    while ((Get-Date) -lt $endTime) {
        $timestamp = (Get-Date).ToString("yyyy-MM-dd HH:mm:ss")
        $procs = Get-Process -Name $ProcessName -ErrorAction SilentlyContinue

        if ($null -eq $procs -or $procs.Count -eq 0) {
            Write-Host "[$timestamp] Proses '$ProcessName' tidak ditemukan, menunggu..." -ForegroundColor Yellow
            Start-Sleep -Seconds $IntervalSec
            continue
        }

        # Ambil proses pertama (jika ada multiple)
        $proc = $procs | Select-Object -First 1

        # CPU Usage: hitung delta TotalProcessorTime
        $pid = $proc.Id
        $curCpuMs = $proc.TotalProcessorTime.TotalMilliseconds
        $cpuPercent = 0

        if ($prevCpuTime.ContainsKey($pid)) {
            $deltaCpuMs   = $curCpuMs - $prevCpuTime[$pid]
            $deltaWallMs  = $IntervalSec * 1000
            $numCores      = [Environment]::ProcessorCount
            $cpuPercent    = [math]::Round(($deltaCpuMs / ($deltaWallMs * $numCores)) * 100, 2)
            if ($cpuPercent -lt 0)   { $cpuPercent = 0 }
            if ($cpuPercent -gt 100) { $cpuPercent = 100 }
        }
        $prevCpuTime[$pid] = $curCpuMs

        # Memory
        $memMB         = [math]::Round($proc.PrivateMemorySize64 / 1MB, 2)
        $workingSetMB  = [math]::Round($proc.WorkingSet64 / 1MB, 2)

        # Tulis ke CSV
        "$timestamp,$cpuPercent,$memMB,$workingSetMB,$($proc.Name)" |
            Out-File -FilePath $OutputFile -Append -Encoding UTF8

        $sampleCount++

        # Tampilkan progress setiap 5 sampel
        if ($sampleCount % 5 -eq 0) {
            Write-Host "[$timestamp] CPU: ${cpuPercent}% | Mem: ${memMB} MB | WS: ${workingSetMB} MB" -ForegroundColor Green
        } else {
            Write-Host "[$timestamp] CPU: ${cpuPercent}% | Mem: ${memMB} MB" -ForegroundColor Gray
        }

        Start-Sleep -Seconds $IntervalSec
    }
} catch {
    if ($_.Exception -is [System.Management.Automation.PipelineStoppedException]) {
        Write-Host "`nMonitor dihentikan (Ctrl+C)" -ForegroundColor Yellow
    } else {
        Write-Host "Error: $_" -ForegroundColor Red
    }
}

# Hitung statistik akhir dari CSV
Write-Host ""
Write-Host "==================================================" -ForegroundColor Cyan
Write-Host "  HASIL MONITORING: $TestName" -ForegroundColor Cyan
Write-Host "==================================================" -ForegroundColor Cyan

$data = Import-Csv -Path $OutputFile
if ($data.Count -gt 0) {
    $cpuValues  = $data | ForEach-Object { [double]$_.CPU_Percent }
    $memValues  = $data | ForEach-Object { [double]$_.Memory_MB }

    $cpuAvg  = [math]::Round(($cpuValues | Measure-Object -Average).Average, 2)
    $cpuMax  = [math]::Round(($cpuValues | Measure-Object -Maximum).Maximum, 2)
    $memAvg  = [math]::Round(($memValues | Measure-Object -Average).Average, 2)
    $memMax  = [math]::Round(($memValues | Measure-Object -Maximum).Maximum, 2)

    Write-Host "  CPU Rata-rata : $cpuAvg %"   -ForegroundColor White
    Write-Host "  CPU Maksimum  : $cpuMax %"   -ForegroundColor White
    Write-Host "  Mem Rata-rata : $memAvg MB"  -ForegroundColor White
    Write-Host "  Mem Maksimum  : $memMax MB"  -ForegroundColor White
    Write-Host "  Total Sampel  : $($data.Count)" -ForegroundColor White

    # Simpan ringkasan ke file terpisah
    $summaryFile = Join-Path $OutputDir "$TestName`_cpu_mem_summary.txt"
    @"
=== CPU & MEMORY SUMMARY: $TestName ===
Proses      : $ProcessName
Sampel      : $($data.Count)
CPU Avg     : $cpuAvg %
CPU Max     : $cpuMax %
Mem Avg     : $memAvg MB
Mem Max     : $memMax MB
"@ | Out-File -FilePath $summaryFile -Encoding UTF8
    Write-Host ""
    Write-Host "  Ringkasan disimpan ke: $summaryFile" -ForegroundColor Green
}

Write-Host "  CSV data    : $OutputFile" -ForegroundColor Green
Write-Host "==================================================" -ForegroundColor Cyan
