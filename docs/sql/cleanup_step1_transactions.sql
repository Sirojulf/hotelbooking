-- ────────────────────────────────────────────────────────────────────────────
-- STEP 1: Hapus transactions (BATCH BESAR + extended timeout)
--
-- TEKAN RUN BERULANG sampai "remaining = 0".
-- Dengan timeout 5 menit + batch 20000, ~5-6 run untuk 100k+ rows.
-- ────────────────────────────────────────────────────────────────────────────

-- Naikkan timeout per query ini ke 5 menit (default 8s)
SET LOCAL statement_timeout = '5min';

DELETE FROM transactions
WHERE id IN (
    SELECT t.id
    FROM transactions t
    WHERE EXISTS (
        SELECT 1 FROM reservations r
        WHERE r.id = t.reservation_id
          AND r.check_in_date >= CURRENT_DATE + INTERVAL '300 days'
    )
    LIMIT 20000
);

SELECT COUNT(*) AS remaining
FROM transactions t
WHERE EXISTS (
    SELECT 1 FROM reservations r
    WHERE r.id = t.reservation_id
      AND r.check_in_date >= CURRENT_DATE + INTERVAL '300 days'
);