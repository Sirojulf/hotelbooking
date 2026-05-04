# Hotel Booking API

REST API untuk manajemen hotel booking: guest, admin, inventory, booking, pembayaran, dan laporan. Aplikasi ini dibangun dengan Go + Echo dan menggunakan Supabase sebagai Auth dan database PostgREST.

## Fitur utama
- Auth guest dan admin (Supabase Auth)
- Pencarian hotel dan detail hotel
- Booking kamar, pembayaran, invoice, pembatalan
- Manajemen inventory hotel (hotel, room type, room, rate, foto)
- Admin user management
- Laporan ringkas (occupancy, ADR, RevPAR, revenue)
- Swagger UI

## Arsitektur singkat
- `handler` menerima request HTTP dan validasi dasar
- `service` mengelola aturan bisnis
- `repository` berinteraksi dengan Supabase (PostgREST)
- `models` berisi struktur data dan enum

## Teknologi
- Go (lihat `go.mod`)
- Echo v4
- Supabase (Auth + PostgREST)
- Viper
- Swaggo (Swagger)

## Struktur proyek
- `cmd/main.go` entrypoint aplikasi
- `internal/handler` HTTP handlers
- `internal/service` business logic
- `internal/repository` akses data Supabase
- `internal/models` struktur data dan enum
- `internal/middleware` auth dan admin guard
- `docs` Swagger JSON/YAML dan SQL constraint

## Konfigurasi
Gunakan file `.env` di root project atau environment variables:
```
SUPABASE_URL=your_supabase_url
SUPABASE_KEY=your_supabase_service_or_anon_key
```

Catatan:
- Server membaca `.env` dari working directory. Jalankan dari root project.
- `SUPABASE_KEY` harus punya akses ke tabel yang digunakan (public schema).

## Menjalankan aplikasi
```
go run ./cmd
```
Server berjalan di `http://localhost:8080`.

Health check:
- `GET /`

Swagger UI:
- `GET /swagger/index.html`

Regenerasi Swagger:
```
go generate ./...
```
atau:
```
swag init -g cmd/main.go -o docs
```

## Autentikasi
Gunakan header:
```
Authorization: Bearer <access_token>
```
Token diperoleh dari endpoint login guest atau admin (Supabase Auth).

Role admin:
- `SuperAdmin` dapat mengelola semua property.
- Admin biasa hanya bisa mengakses property miliknya.

## Endpoint API (ringkas)
Base path: `/api/v1`

Public:
- `POST /auth/guest/register`
- `POST /auth/guest/login`
- `POST /auth/admin/login`
- `GET /hotels?city=Jakarta`
- `GET /hotels/:id`
- `GET /rooms/:room_id/availability?check_in=YYYY-MM-DD&check_out=YYYY-MM-DD`

Guest (BearerAuth):
- `GET /guests/me`
- `GET /guests/bookings`
- `POST /guests/bookings`
- `POST /guests/bookings/:id/pay`
- `POST /guests/bookings/:id/cancel`
- `GET /guests/bookings/:id/invoice`

Admin (BearerAuth + AdminOnly):
- `POST /admin/hotels`
- `GET /admin/hotels`
- `PUT /admin/hotels/:id`
- `DELETE /admin/hotels/:id`
- `POST /admin/room-types`
- `PUT /admin/room-types/:id`
- `DELETE /admin/room-types/:id`
- `GET /admin/room-types`
- `POST /admin/rooms`
- `PUT /admin/rooms/:id`
- `DELETE /admin/rooms/:id`
- `GET /admin/rooms`
- `POST /admin/rooms/:room_id/rates`
- `GET /admin/rooms/:room_id/rates`
- `POST /admin/hotels/:property_id/photos`
- `GET /admin/hotels/:property_id/photos`
- `DELETE /admin/photos/property/:id`
- `POST /admin/room-photos`
- `GET /admin/room-photos`
- `DELETE /admin/photos/room/:id`
- `POST /admin/users`
- `GET /admin/users`
- `PUT /admin/users/:id`
- `POST /admin/users/:id/activate`
- `POST /admin/users/:id/deactivate`
- `GET /admin/bookings`
- `PUT /admin/bookings/:id/status`
- `GET /admin/reports/summary`

Detail request/response lengkap tersedia di Swagger UI.

## Model data ringkas
Tabel utama:
- `guests`: profil guest
- `admin`: user admin
- `properties`: hotel
- `room_types`: tipe kamar
- `rooms`: unit kamar
- `room_rates`: harga per tanggal
- `bookings`: data booking
- `payments`: status pembayaran
- `invoices`: invoice booking
- `property_photos`, `room_photos`

Enum penting (lihat `internal/models/enums.go`):
- `GuestType`: `Adult`, `Child`
- `Gender`: `Male`, `Female`
- `VIPStatus`: `Bronze`, `Silver`, `Gold`, `Platinum`
- `BookingStatus`: `New`, `Confirmed`, `Cancelled`, `CheckedIn`, `CheckedOut`, `NoShow`
- `RoomStatus`: `Available`, `Occupied`, `OutOfOrder`
- `HousekeepingStatus`: `Clean`, `Dirty`, `Inspected`, `Pickup`, `OutOfOrder`, `OutOfService`
- `PaymentStatus`: `Pending`, `Paid`, `Refunded`

## Aturan booking dan rate
- Availability mempertimbangkan status kamar dan housekeeping.
- `room_rates` mendukung rate per tanggal:
  - `linear_rate` untuk harga tetap.
  - `non_linear_rate` untuk struktur JSON (misal per occupancy).
  - `stop_sell`, `close_on_arrival`, `close_on_departure`, `min_nights`, `max_nights`.
- Total harga dihitung dari rate per malam, default ke base price room type.

## Database constraint
Untuk mencegah overlapping booking per room:
```
docs/db_constraints.sql
```
Jalankan di Supabase SQL editor setelah data bersih.

## Catatan tambahan
- Aplikasi ini mengandalkan Supabase Auth untuk login dan validasi token.
- Untuk data seed dan struktur tabel, sesuaikan dengan model di `internal/models`.
