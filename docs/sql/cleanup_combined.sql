-- ────────────────────────────────────────────────────────────────────────────
-- cleanup_combined.sql
-- Hapus transactions + reservations sekaligus per batch (10.000 reservations).
-- Tidak akan timeout karena pakai SET LOCAL statement_timeout.
--
-- TEKAN RUN BERULANG sampai "res_deleted = 0" (atau "reservations_left = 0").
-- ────────────────────────────────────────────────────────────────────────────

SET LOCAL statement_timeout = '5min';

WITH target_ids AS (
    SELECT id FROM reservations
    WHERE check_in_date >= CURRENT_DATE + INTERVAL '300 days'
    LIMIT 10000
),
deleted_tx AS (
    DELETE FROM transactions
    WHERE reservation_id IN (SELECT id FROM target_ids)
    RETURNING 1
),
deleted_res AS (
    DELETE FROM reservations
    WHERE id IN (SELECT id FROM target_ids)
    RETURNING 1
)
SELECT
    (SELECT COUNT(*) FROM deleted_tx)  AS tx_deleted,
    (SELECT COUNT(*) FROM deleted_res) AS res_deleted;

-- Cek sisa
SELECT
    (SELECT COUNT(*) FROM reservations WHERE check_in_date >= CURRENT_DATE + INTERVAL '300 days') AS reservations_left,
    (SELECT COUNT(*) FROM transactions t
        WHERE EXISTS (SELECT 1 FROM reservations r
                      WHERE r.id = t.reservation_id
                      AND r.check_in_date >= CURRENT_DATE + INTERVAL '300 days')) AS orphan_transactions_left;