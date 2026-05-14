-- ────────────────────────────────────────────────────────────────────────────
-- create_reservation_atomic (v2 — pakai exclusion constraint)
-- ────────────────────────────────────────────────────────────────────────────
-- Versi baru memanfaatkan exclusion constraint untuk atomicity di DB level.
-- Tidak ada FOR UPDATE — lebih cepat di high concurrency.
-- Tidak ada manual overlap check — DB sendiri yang enforce.
-- Race condition impossible — exclusion constraint atomically checked.
--
-- PRASYARAT: jalankan add_exclusion_constraint.sql DULU.
-- ────────────────────────────────────────────────────────────────────────────

CREATE OR REPLACE FUNCTION create_reservation_atomic(
    p_id               UUID,
    p_hotel_id         UUID,
    p_guest_id         UUID,
    p_room_id          UUID,
    p_check_in         DATE,
    p_check_out        DATE,
    p_total_price      NUMERIC,
    p_booking_source   TEXT,
    p_special_requests TEXT,
    p_payment_method   TEXT,
    p_tx_id            UUID,
    p_tx_description   TEXT
) RETURNS JSONB
LANGUAGE plpgsql
AS $$
DECLARE
    v_reservation reservations%ROWTYPE;
    v_transaction transactions%ROWTYPE;
BEGIN
    -- INSERT akan FAIL otomatis kalau ada overlap (exclusion constraint).
    INSERT INTO reservations (
        id, hotel_id, guest_id, room_id,
        check_in_date, check_out_date, total_price,
        payment_status, booking_source, special_requests, payment_method,
        created_at
    ) VALUES (
        p_id, p_hotel_id, p_guest_id, p_room_id,
        p_check_in, p_check_out, p_total_price,
        'pending',
        NULLIF(p_booking_source, '')::booking_source,
        p_special_requests,
        p_payment_method,
        NOW()
    ) RETURNING * INTO v_reservation;

    INSERT INTO transactions (
        id, reservation_id, description, amount, type, created_at
    ) VALUES (
        p_tx_id, p_id, p_tx_description, p_total_price,
        'charge',
        NOW()
    ) RETURNING * INTO v_transaction;

    RETURN jsonb_build_object(
        'reservation', row_to_json(v_reservation),
        'transaction', row_to_json(v_transaction)
    );

EXCEPTION
    -- Exclusion constraint violation → conflict (atomic, no race window)
    WHEN exclusion_violation THEN
        RAISE EXCEPTION 'ROOM_CONFLICT' USING ERRCODE = 'P0001';
END;
$$;

GRANT EXECUTE ON FUNCTION create_reservation_atomic(
    UUID, UUID, UUID, UUID, DATE, DATE, NUMERIC, TEXT, TEXT, TEXT, UUID, TEXT
) TO anon, authenticated;