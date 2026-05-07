# Panduan Load Testing — hotelbooking (Go) vs RoomMasterb (Node.js)

Panduan ini menjelaskan langkah-langkah lengkap untuk menjalankan Load Testing, Spike Testing, dan Stress Testing menggunakan k6, serta cara memantau CPU/Memory server.

---

## Daftar Isi

1. [Prasyarat](#1-prasyarat)
2. [Struktur File](#2-struktur-file)
3. [Konfigurasi Awal](#3-konfigurasi-awal)
4. [Quick Test (Verifikasi)](#4-quick-test-verifikasi)
5. [Load Testing](#5-load-testing)
6. [Spike Testing](#6-spike-testing)
7. [Stress Testing](#7-stress-testing)
8. [Monitor CPU & Memory](#8-monitor-cpu--memory)
9. [Urutan Lengkap Satu Sesi Pengujian](#9-urutan-lengkap-satu-sesi-pengujian)
10. [Analisis Hasil](#10-analisis-hasil)
11. [Testing RoomMasterb (Node.js)](#11-testing-roomasterb-nodejs)
12. [Perbandingan Hasil](#12-perbandingan-hasil)
13. [Troubleshooting](#13-troubleshooting)

---

## 1. Prasyarat

Pastikan semua tools berikut sudah terinstall:

| Tool | Versi | Cek Instalasi |
|------|-------|---------------|
| Go | ≥ 1.21 | `go version` |
| k6 | ≥ 1.0 | `k6 version` |
| Python | ≥ 3.8 | `python --version` |
| PowerShell | ≥ 7 | `$PSVersionTable.PSVersion` |

Install dependensi Python untuk analisis:
```bash
pip install matplotlib pandas
```

---

## 2. Struktur File

```
k6/
├── config.js                  # Konfigurasi URL, kredensial, room IDs
├── load_test.js               # Skenario Load Testing
├── spike_test.js              # Skenario Spike Testing
├── stress_test.js             # Skenario Stress Testing
├── flow/
│   ├── hotelbooking.js        # Flow booking untuk Go API
│   └── roommaster.js          # Flow booking untuk Node.js API
├── monitor.ps1                # Monitor CPU/Memory (Windows)
├── analyze.py                 # Analisis & chart hasil
└── results/                   # Output JSON, CSV, chart (auto-dibuat)
```

---

## 3. Konfigurasi Awal

### 3.1 Edit `config.js`

Buka `k6/config.js` dan sesuaikan:

```js
export const HOTELBOOKING = {
  baseURL: "http://localhost:8080/api/v1",
  email: "k6test@mail.com",      // akun guest yang sudah terdaftar
  password: "password123",
  hotelId: "a0000000-0000-0000-0000-000000000001",  // UUID hotel dari DB
  roomIds: [
    "a0000000-0000-0000-0000-000000000200",  // minimal 5 room
    "a0000000-0000-0000-0000-000000000201",
    "a0000000-0000-0000-0000-000000000202",
    "a0000000-0000-0000-0000-000000000203",
    "a0000000-0000-0000-0000-000000000204",
  ],
};
```

> **Penting:** Pastikan akun `k6test@mail.com` sudah terdaftar via `POST /api/v1/auth/guest/register` dan 5 room sudah ada di database.

### 3.2 Buat akun guest untuk testing (jika belum ada)

```bash
curl -X POST http://localhost:8080/api/v1/auth/guest/register \
  -H "Content-Type: application/json" \
  -d '{"email":"k6test@mail.com","password":"password123","full_name":"K6 Test User","hotel_id":"a0000000-0000-0000-0000-000000000001"}'
```

---

## 4. Quick Test (Verifikasi)

Sebelum menjalankan test penuh, verifikasi bahwa flow berjalan benar dengan 1 VU dan 1 iterasi.

### Terminal 1 — Jalankan server Go

```powershell
# Dari root project
go run cmd/main.go
```

### Terminal 2 — Quick test

```powershell
# Dari root project
cd k6
k6 run --vus 1 --iterations 1 .\flow\hotelbooking.js
```

**Hasil yang diharapkan:**

```
✓ login 200
✓ availability 200
✓ reservation 201
✓ cancel 200
```

Jika semua centang hijau, lanjut ke step berikutnya.

---

## 5. Load Testing

**Tujuan:** Mengukur kinerja sistem pada beban normal yang stabil (sustained load).

**Skenario:**
- 0 → 20 VU dalam 1 menit (ramp-up)
- 20 VU selama 3 menit (steady state)
- 20 → 0 VU dalam 1 menit (ramp-down)
- **Total durasi: ±5 menit**

**Threshold:**
- HTTP P95 < 3000ms
- Error rate < 5%
- Booking success > 90%

### Jalankan Load Test

```powershell
# Terminal 1 — Server Go sudah berjalan

# Terminal 2 — Monitor (opsional tapi dianjurkan)
.\monitor.ps1 -TestName "load_hotelbooking" -ProcessName "main" -DurationSec 360

# Terminal 3 — Load Test
k6 run .\load_test.js
```

> **Catatan:** Jalankan monitor SEBELUM k6 dimulai agar data CPU/Memory tercatat lengkap.
> Jangan pakai flag `--out json=` — file JSON summary sudah ditulis otomatis oleh `handleSummary`.

### Output yang dihasilkan

- `k6/results/load_hotelbooking.json` — data metrik summary (ditulis otomatis)
- `k6/results/load_hotelbooking_monitor.csv` — data CPU/Memory per detik
- `k6/results/load_hotelbooking_cpu_mem_summary.txt` — ringkasan CPU/Memory

---

## 6. Spike Testing

**Tujuan:** Mengukur kemampuan sistem menghadapi lonjakan traffic mendadak.

**Skenario:**
- 0 → 5 VU (baseline, 30 detik)
- Tahan 5 VU (30 detik)
- **5 → 50 VU mendadak (10 detik) ← ini spike-nya**
- Tahan 50 VU (1 menit)
- 50 → 5 VU (10 detik)
- Recovery check (30 detik)
- **Total durasi: ±3 menit**

**Threshold:**
- HTTP P95 < 5000ms (lebih longgar karena spike)
- Error rate < 10%
- Booking success > 80%

### Jalankan Spike Test

```powershell
# Terminal 1 — Server Go sudah berjalan

# Terminal 2 — Monitor
.\monitor.ps1 -TestName "spike_hotelbooking" -ProcessName "main" -DurationSec 240

# Terminal 3 — Spike Test
k6 run .\spike_test.js
```

---

## 7. Stress Testing

**Tujuan:** Menemukan breaking point sistem dengan menaikkan beban secara bertahap.

**Skenario:**
- 0 → 10 VU (1 menit) — normal
- 10 → 20 VU (1 menit) — di atas normal
- 20 → 40 VU (1 menit) — mulai berat
- 40 → 60 VU (1 menit) — stress
- 60 → 80 VU (1 menit) — heavy stress
- 80 → 100 VU (1 menit) — extreme
- 100 → 0 VU (1 menit) — recovery
- **Total durasi: ±7 menit**

**Threshold (lebih longgar — kita ingin melihat di mana sistem patah):**
- HTTP P95 < 10000ms
- Error rate < 20%
- Booking success > 70%

### Jalankan Stress Test

```powershell
# Terminal 1 — Server Go sudah berjalan

# Terminal 2 — Monitor
.\monitor.ps1 -TestName "stress_hotelbooking" -ProcessName "main" -DurationSec 480

# Terminal 3 — Stress Test
k6 run .\stress_test.js
```

---

## 8. Monitor CPU & Memory

Monitor berjalan di terminal terpisah dan mencatat CPU/Memory proses Go server setiap 1 detik.

### Parameter monitor.ps1

| Parameter | Default | Keterangan |
|-----------|---------|------------|
| `-TestName` | `"monitor"` | Nama file output |
| `-ProcessName` | `"main"` | Nama proses yang dipantau |
| `-IntervalSec` | `1` | Interval sampling (detik) |
| `-DurationSec` | `600` | Durasi maksimum (detik) |

### Contoh penggunaan

```powershell
# Untuk load test (360 detik = 5 menit + buffer)
.\monitor.ps1 -TestName "load_hotelbooking" -ProcessName "main" -DurationSec 360

# Tekan Ctrl+C setelah k6 selesai untuk hentikan monitor lebih awal
```

### Output monitor

Monitor otomatis mencetak ringkasan saat dihentikan:

```
==================================================
  HASIL MONITORING: load_hotelbooking
==================================================
  CPU Rata-rata : 12.5 %
  CPU Maksimum  : 48.3 %
  Mem Rata-rata : 24.7 MB
  Mem Maksimum  : 31.2 MB
  Total Sampel  : 302
==================================================
```

---

## 9. Urutan Lengkap Satu Sesi Pengujian

Lakukan ini secara berurutan. Setiap test butuh jeda untuk server recovery.

```
Jeda minimal 2 menit antar test agar hasil tidak saling mempengaruhi.
```

### Step-by-step

**Step 1 — Persiapan**
```powershell
# Pastikan direktori hasil ada
New-Item -ItemType Directory -Force -Path .\results
```

**Step 2 — Jalankan server (Terminal 1, biarkan berjalan)**
```powershell
go run cmd/main.go
```

**Step 3 — Quick test (Terminal 2)**
```powershell
k6 run --vus 1 --iterations 1 .\flow\hotelbooking.js
```

**Step 4 — Load Test (buka Terminal 2 & 3)**
```powershell
# Terminal 2
.\monitor.ps1 -TestName "load_hotelbooking" -ProcessName "main" -DurationSec 360

# Terminal 3
k6 run .\load_test.js

# Setelah selesai: Ctrl+C pada Terminal 2, tunggu 2 menit
```

**Step 5 — Spike Test**
```powershell
# Terminal 2
.\monitor.ps1 -TestName "spike_hotelbooking" -ProcessName "main" -DurationSec 240

# Terminal 3
k6 run .\spike_test.js

# Setelah selesai: Ctrl+C pada Terminal 2, tunggu 2 menit
```

**Step 6 — Stress Test**
```powershell
# Terminal 2
.\monitor.ps1 -TestName "stress_hotelbooking" -ProcessName "main" -DurationSec 480

# Terminal 3
k6 run .\stress_test.js

# Setelah selesai: Ctrl+C pada Terminal 2
```

---

## 10. Analisis Hasil

Setelah semua test selesai, jalankan script analisis dari **root project**:

```powershell
python k6/analyze.py
```

Script ini akan menghasilkan:

| File | Keterangan |
|------|------------|
| `results/comparison_table.txt` | Tabel perbandingan teks |
| `results/chart_success_requests.png` | Chart total HTTP requests |
| `results/chart_throughput.png` | Chart req/detik |
| `results/chart_error_rate.png` | Chart error rate |
| `results/chart_response_p95.png` | Chart response time P95 |
| `results/chart_booking_success.png` | Chart booking success rate |

### Contoh tabel hasil

```
==========================================================================================
  TABEL PERBANDINGAN KINERJA: hotelbooking (Go) vs RoomMasterb (Node.js)
==========================================================================================

  [ LOAD TESTING ]
  ------------------------------------------------------------------------------------------
  Metrik                              Go (hotelbooking)      Node.js (RoomMasterb)
  ------------------------------------------------------------------------------------------
  Total HTTP Requests                          1200                    980
  Req / detik                               4.00 rps               3.26 rps
  Error Rate                                0.83%                   2.14%
  Booking Success Rate                       98.5%                  91.2%
  ...
```

---

## 11. Testing RoomMasterb (Node.js)

### 11.1 Clone & jalankan RoomMasterb

```bash
git clone https://github.com/Sirojulf/RoomMasterb.git
cd RoomMasterb
npm install
npm start   # atau sesuai dokumentasi repo
```

### 11.2 Isi konfigurasi RoomMasterb di `config.js`

```js
export const ROOMMASTER = {
  baseURL: "http://localhost:3000/api",   // sesuaikan port
  email: "fo@hotel.com",
  password: "password123",
  roomIds: [
    "UUID_ROOM_1",   // ambil dari DB RoomMasterb
    "UUID_ROOM_2",
    "UUID_ROOM_3",
    "UUID_ROOM_4",
    "UUID_ROOM_5",
  ],
};
```

### 11.3 Update flow/roommaster.js

Sesuaikan endpoint dan format request/response di `k6/flow/roommaster.js` dengan API RoomMasterb.

### 11.4 Ganti import di test files

Di `load_test.js`, `spike_test.js`, `stress_test.js` — ganti baris import:

```js
// Komen baris ini:
// import { runBookingFlow } from './flow/hotelbooking.js';

// Aktifkan baris ini:
import { runBookingFlow } from './flow/roommaster.js';
```

### 11.5 Jalankan test untuk RoomMasterb

```powershell
# Quick test
k6 run --vus 1 --iterations 1 .\flow\roommaster.js

# Load Test
.\monitor.ps1 -TestName "load_roommaster" -ProcessName "node" -DurationSec 360
k6 run .\load_test.js

# Spike Test
.\monitor.ps1 -TestName "spike_roommaster" -ProcessName "node" -DurationSec 240
k6 run .\spike_test.js

# Stress Test
.\monitor.ps1 -TestName "stress_roommaster" -ProcessName "node" -DurationSec 480
k6 run .\stress_test.js
```

> **Catatan:** Ganti `-ProcessName "node"` dengan nama proses yang muncul di Task Manager saat menjalankan Node.js.

---

## 12. Perbandingan Hasil

Setelah semua 6 file JSON tersedia di `k6/results/`:

```
results/
├── load_hotelbooking.json
├── spike_hotelbooking.json
├── stress_hotelbooking.json
├── load_roommaster.json
├── spike_roommaster.json
└── stress_roommaster.json
```

Jalankan analisis final:

```powershell
python k6/analyze.py
```

Buka file chart yang dihasilkan untuk dimasukkan ke laporan tugas akhir.

---

## 13. Troubleshooting

### Error: `function 'default' not found in exports`
Pastikan menjalankan file test yang benar, bukan file flow:
```powershell
# Benar
k6 run .\load_test.js

# Salah (file flow tidak punya export default, kecuali untuk quick test)
k6 run .\flow\hotelbooking.js
```

### Error: `422 Unprocessable Entity` saat buat reservasi
- Pastikan room ID di `config.js` ada di database
- Pastikan room tersebut status `available` dan `clean`
- Cek apakah tanggal check-in/out sudah lewat

### Error: `401 Unauthorized`
- Token JWT expired — k6 login ulang di setiap iterasi secara otomatis
- Pastikan kredensial di `config.js` sesuai dengan akun di database

### Monitor tidak mendeteksi proses
```powershell
# Cari nama proses yang benar
Get-Process | Where-Object { $_.Name -like "*main*" -or $_.Name -like "*node*" }
```

### k6 tidak ditemukan
```powershell
# Install k6 via winget
winget install k6 --source winget

# Atau via Chocolatey
choco install k6
```

### Hasil JSON kosong / tidak ada file output
Pastikan direktori `results/` sudah ada sebelum menjalankan k6:
```powershell
New-Item -ItemType Directory -Force -Path .\results
```
