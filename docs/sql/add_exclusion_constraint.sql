-- ────────────────────────────────────────────────────────────────────────────
-- add_exclusion_constraint.sql
-- Tambah PostgreSQL exclusion constraint untuk prevent overlapping reservations
-- di DB level. Ini menggantikan FOR UPDATE locking — lebih cepat, lebih kuat.
--
-- Setelah constraint ini ada:
--   - INSERT yang akan create overlapping reservation otomatis DITOLAK
--   - Error code: 23P01 (exclusion_violation)
--   - Tidak perlu manual overlap check di Go atau RPC
--   - Tidak ada race condition (constraint dijamin atomic oleh PostgreSQL)
--
-- PERSYARATAN:
--   - Tidak ada overlapping reservations existing di DB (jalankan cleanup dulu)
--   - btree_gist extension tersedia (default ada di Supabase)
-- ────────────────────────────────────────────────────────────────────────────

-- 1. Enable extension (idempotent — aman dijalankan berkali-kali)
CREATE EXTENSION IF NOT EXISTS btree_gist;

-- 2. Drop dulu kalau constraint sudah pernah dibuat (aman dijalankan ulang)
ALTER TABLE reservations
    DROP CONSTRAINT IF EXISTS no_overlapping_active_reservations;

-- 3. Add exclusion constraint:
--    Tolak INSERT/UPDATE yang menyebabkan 2 reservation aktif
--    punya room_id sama DAN rentang tanggal overlapping.
ALTER TABLE reservations
ADD CONSTRAINT no_overlapping_active_reservations
EXCLUDE USING gist (
    room_id WITH =,
    daterange(check_in_date, check_out_date, '[)') WITH &&
)
WHERE (payment_status <> 'cancelled');

-- 4. Verifikasi constraint sudah aktif
SELECT
    conname        AS constraint_name,
    pg_get_constraintdef(oid) AS definition
FROM pg_constraint
WHERE conrelid = 'reservations'::regclass
  AND contype = 'x';