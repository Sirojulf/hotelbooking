-- ────────────────────────────────────────────────────────────────────────────
-- STEP 2: Hapus reservations (BATCH BESAR + extended timeout)
--
-- ⚠️  HANYA jalankan setelah STEP 1 selesai (transactions remaining = 0)
-- TEKAN RUN BERULANG sampai "remaining = 0".
-- ────────────────────────────────────────────────────────────────────────────

SET LOCAL statement_timeout = '5min';

DELETE FROM reservations
WHERE id IN (
    SELECT id FROM reservations
    WHERE check_in_date >= CURRENT_DATE + INTERVAL '300 days'
    LIMIT 20000
);

SELECT COUNT(*) AS remaining
FROM reservations
WHERE check_in_date >= CURRENT_DATE + INTERVAL '300 days';