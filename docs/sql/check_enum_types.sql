-- ────────────────────────────────────────────────────────────────────────────
-- Cek nama enum types yang ada di schema Supabase kamu.
-- Jalankan ini DULU sebelum deploy create_reservation_atomic.sql
-- supaya tahu kalau ada nama enum yang berbeda.
-- ────────────────────────────────────────────────────────────────────────────

-- 1. List semua enum types & nilai-nilainya
SELECT
    t.typname  AS enum_name,
    string_agg(e.enumlabel, ', ' ORDER BY e.enumsortorder) AS values
FROM pg_type t
JOIN pg_enum e ON t.oid = e.enumtypid
JOIN pg_namespace n ON n.oid = t.typnamespace
WHERE n.nspname = 'public'
GROUP BY t.typname
ORDER BY t.typname;

-- 2. Cek tipe kolom tabel reservations & transactions
SELECT
    table_name,
    column_name,
    data_type,
    udt_name      -- udt_name = nama enum kalau columnnya enum
FROM information_schema.columns
WHERE table_name IN ('reservations', 'transactions')
  AND table_schema = 'public'
ORDER BY table_name, ordinal_position;