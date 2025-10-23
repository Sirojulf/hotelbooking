#  Guest Login

## Overview
Endpoint ini digunakan untuk melakukan login tamu (**guest**) menggunakan **email atau nomor telepon** yang telah terdaftar pada sistem Hotel Booking.  
Autentikasi ditangani melalui **Supabase Auth** menggunakan kombinasi email/password atau phone/password.

---

##  Endpoint
**POST** `/api/v1/guest/login`

---

##  Request Body

### Format
```json
{
  "login": "string (email or phone number)",
  "password": "string"
}
```

### Login berhasil
```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI...",
  "token_type": "bearer",
  "expires_in": 3600,
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI...",
  "user": {
    "id": "a12b34cd-5678-ef90-gh12-ijk345lm678n",
    "aud": "authenticated",
    "email": "guest@example.com",
    "phone": "081234567123",
    "role": "authenticated",
    "created_at": "2025-10-21T07:12:45.000Z",
    "last_sign_in_at": "2025-10-21T07:13:00.000Z"
  }
}

```

###  400 Bad Request — Body Tidak Valid
``` json
{
  "error": "Invalid request body"
}
```
