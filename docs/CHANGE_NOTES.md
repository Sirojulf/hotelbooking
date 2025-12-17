# Ringkasan Perubahan Terbaru

Commit terakhir memperkuat proses inisialisasi klien Supabase dan menambahkan perlindungan di lapisan repository serta service agar aplikasi gagal dengan pesan yang lebih jelas ketika konfigurasi Supabase belum benar.

## Detail Utama
- `internal/config/supabase.go`: fungsi `ConnectSupabase` kini memvalidasi variabel `SUPABASE_URL` dan `SUPABASE_KEY`, meneruskan kesalahan dari `supabase.NewClient`, dan mengembalikan error ke pemanggil agar startup dapat ditangani secara eksplisit.
- `cmd/API/main.go`: startup server sekarang memeriksa error dari `ConnectSupabase` dan akan menampilkan pesan fatal jika koneksi Supabase gagal.
- Repository (`internal/repository/*.go`) dan service login/registrasi (`internal/service/*.go`) menambahkan pengecekan `config.SupabaseClient == nil` sebelum melakukan operasi database sehingga menghindari panic akibat klien belum terinisialisasi.

Perubahan ini bertujuan memberikan pengalaman debugging yang lebih jelas dan mencegah akses ke Supabase ketika kredensial belum tersedia.
