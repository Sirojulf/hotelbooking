// ─── Konfigurasi Target ───────────────────────────────────────────────────────
// Ganti nilai placeholder dengan data REAL dari environment kamu

export const HOTELBOOKING = {
  baseURL: "http://localhost:8080/api/v1",
  // Akun guest yang sudah ada di DB (buat dulu via POST /auth/guest/register)
  email: "k6test@mail.com",
  password: "password123",
  hotelId: "a0000000-0000-0000-0000-000000000001",
  // Sediakan minimal 5 room agar VU tidak rebutan slot tanggal yang sama
  roomIds: [
    "a0000000-0000-0000-0000-000000000200", // Room 101 - Standard
    "a0000000-0000-0000-0000-000000000201", // Room 102 - Deluxe
    "a0000000-0000-0000-0000-000000000202", // Room 103 - Suite
    "a0000000-0000-0000-0000-000000000203", // Room 104 - Family
    "a0000000-0000-0000-0000-000000000204", // Room 105 - Executive
  ],
};

export const ROOMMASTER = {
  baseURL: "http://localhost:3000/api",
  email: "k6test@roommaster.test",
  password: "K6testPassword123!",
  roomIds: [
    "b0000000-0000-0000-0000-000000000001",
    "b0000000-0000-0000-0000-000000000002",
    "b0000000-0000-0000-0000-000000000003",
    "b0000000-0000-0000-0000-000000000004",
    "b0000000-0000-0000-0000-000000000005",
  ],
};

// ─── Helper: tanggal check-in/out per VU + iterasi ───────────────────────────
// Setiap kombinasi (vuId, iter) mendapat slot tanggal unik agar tidak pernah
// bentrok meski cancel gagal di iterasi sebelumnya.
// Rumus: offset = 365 + vuId*100 + iter
//   VU 0  iter 0  → hari ke-365
//   VU 0  iter 1  → hari ke-366
//   VU 1  iter 0  → hari ke-465
//   VU 19 iter 0  → hari ke-2265 (±6 tahun, masih valid di Supabase)
export function getCheckInOut(vuId = 0, iter = 0) {
  const base = new Date();
  base.setDate(base.getDate() + 365 + vuId * 100 + iter);
  const checkIn = base.toISOString().split("T")[0];
  const out = new Date(base);
  out.setDate(out.getDate() + 2);
  const checkOut = out.toISOString().split("T")[0];
  return { checkIn, checkOut };
}

// Round-robin room assignment
export function pickRoom(roomIds, vuId = 0) {
  return roomIds[vuId % roomIds.length];
}
