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
  // Sesuaikan dengan port RoomMasterb kamu
  baseURL: "http://localhost:3000/api",
  email: "fo@hotel.com",
  password: "password123",
  // UUID room dari database RoomMasterb
  roomIds: [
    "GANTI_ROOM_UUID_1",
    "GANTI_ROOM_UUID_2",
    "GANTI_ROOM_UUID_3",
    "GANTI_ROOM_UUID_4",
    "GANTI_ROOM_UUID_5",
  ],
};

// ─── Helper: tanggal check-in/out per VU ─────────────────────────────────────
// Setiap VU dapat slot tanggal unik agar tidak bentrok reservation
export function getCheckInOut(vuId = 0) {
  const base = new Date();
  // Mulai 90 hari dari sekarang, tiap VU dapat slot berbeda
  base.setDate(base.getDate() + 90 + (vuId % 60));
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
