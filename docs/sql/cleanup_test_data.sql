-- ────────────────────────────────────────────────────────────────────────────
-- cleanup_test_data.sql
-- Hapus orphan k6 test data. Karena FK constraint, urutan-nya HARUS:
--   1. Hapus SEMUA transactions yang reference target reservations
--   2. BARU hapus reservations
--
-- Statement timeout Supabase = 8s (free tier). Pakai batch kecil + run berulang.
-- ────────────────────────────────────────────────────────────────────────────

-- Cek total yang harus dihapus
SELECT
    (SELECT COUNT(*) FROM reservations WHERE check_in_date >= CURRENT_DATE + INTERVAL '300 days') AS reservations_to_delete,
    (SELECT COUNT(*) FROM transactions t
        WHERE EXISTS (SELECT 1 FROM reservations r
                      WHERE r.id = t.reservation_id
                      AND r.check_in_date >= CURRENT_DATE + INTERVAL '300 days')) AS transactions_to_delete;