# Keamanan & RBAC

> Bagaimana login, sesi, dan pembatasan peran bekerja hari ini, celah apa yang masih ada, dan daftar periksa sebelum MedPath dipakai dengan data rumah sakit sungguhan.

## Peran yang dikenali backend

| Kode | Label | Akun demo |
|---|---|---|
| `SUPER_ADMIN` | Super Admin | ✓ |
| `SYSTEM_ADMIN` | System Admin | — |
| `ADMIN_RS` | Admin RS | — |
| `KOORDINATOR_CP` | Koordinator CP | ✓ |
| `TIM_CP` | Tim CP | ✓ |
| `DPJP` | DPJP | ✓ |
| `DOKTER` | Dokter | — |
| `APOTEKER` | Apoteker | ✓ |
| `KOMITE_MEDIK` | Komite Medik | ✓ |
| `DIREKTUR` | Direktur | ✓ |

## Tiga lapis pembatasan

| Lapis | Di mana | Contoh |
|---|---|---|
| 1. Autentikasi | `middleware.go` → `auth` | tanpa token valid → `401` |
| 2. Gerbang peran per endpoint | `hanyaSuperAdmin` | `/panduan` → `403` untuk selain SUPER_ADMIN |
| 3. Aturan domain per aksi | `store` (`transisiDari`, `pengubahIsi`, `bolehPetaKode`) | DPJP mencoba SAHKAN → `422` |

Frontend juga menyembunyikan menu/tombol (guard `meta.admin`, `meta.superAdmin`, `boleh_ubah`, `aksi_tersedia`) — tetapi itu **hanya kosmetik**. Setiap batas keamanan **wajib** ditegakkan di backend.

## Akses menu per peran

| Menu | Siapa |
|---|---|
| Dashboard, CP Library, Detail CP, Approval, CP Aktif, Evaluasi, Dokumen, Padanan (lihat) | semua yang login |
| Edit padanan KPTL/SNOMED | TIM_CP, KOORDINATOR_CP, ADMIN_RS, SYSTEM_ADMIN, SUPER_ADMIN |
| Pengaturan | ADMIN_RS, SYSTEM_ADMIN, SUPER_ADMIN |
| Panduan Lengkap Super Admin | **SUPER_ADMIN saja** (UI + server) |

## Alur login & sesi

1. `POST /auth/login` memverifikasi email dan hash sandi terhadap snapshot pengguna aktif dari database.
2. Bila cocok, server membuat token acak 24 byte (`crypto/rand`, hex) dan menyimpan hash token dan masa berlaku 24 jam di `auth_sessions`. Demo lokal eksplisit memakai peta memori.
3. Frontend menyimpan token di `localStorage` dan mengirimnya sebagai `Bearer` di setiap request.
4. `POST /auth/logout` mencabut sesi dari penyimpanan sehingga berlaku lintas instance.

## Keterbatasan saat ini (WAJIB dibaca)

| # | Kondisi sekarang | Risiko | Perbaikan |
|---|---|---|---|
| 1 | Hash sandi PBKDF2-HMAC-SHA256, salt acak, 600.000 iterasi | akun demo tetap dikenal sampai dinonaktifkan | buat Super Admin non-demo lalu nonaktifkan demo dari akun tersebut |
| 2 | Sesi PostgreSQL persisten | hash token, kedaluwarsa 24 jam, logout lintas instance | bersihkan sesi kedaluwarsa secara terjadwal |
| 3 | Token di `localStorage` | rentan dicuri jika ada XSS | pertimbangkan cookie `HttpOnly; Secure; SameSite=Strict` |
| 4 | CORS memakai daftar origin eksplisit | origin lain tidak mendapat izin CORS | atur ALLOWED_ORIGINS untuk frontend lintas origin |
| 5 | Tidak ada pembatasan laju login | tebak-sandi tanpa batas | rate limit per IP/email, kunci sementara |
| 6 | Tidak ada HTTPS di server Go | token lewat jalur polos jika diekspos langsung | pasang di belakang reverse proxy TLS (Caddy/Nginx) |
| 7 | Alasan KEMBALIKAN divalidasi server setelah trim | alasan kosong ditolak 422 | pertahankan pengujian regresi |
| 8 | Transaksi per request | perubahan dan audit commit bersama | kegagalan database menjadi 503 |

## Data pasien

Seluruh data di MedPath saat ini **fiktif dan tanpa identitas pasien**. CP bekerja di level agregat (episode per kode INA-CBG/ICD), bukan rekam medis individual. Saat impor klaim riil dibangun, pastikan hanya field agregat yang dimuat (kode grouper, severity, LOS, kelas, tarif, diagnosis) — **tanpa** nama, NIK, nomor RM, atau nomor SEP yang bisa ditelusuri.

## Rahasia & repositori

- `.env` dan file `*.local` di-*ignore* git.
- Repo `official-spkd/MEDPATH_SPKD` bersifat **public**.
- Verifikasi akun GitHub aktif dan izin repository sebelum push. Jangan simpan token dalam kode atau log.

## Daftar periksa sebelum produksi

- [x] Sandi di-hash; form login produksi tidak menampilkan akun demo
- [ ] Akun demo dinonaktifkan setelah akses Super Admin non-demo terverifikasi
- [x] Sesi persisten dengan kedaluwarsa
- [ ] CORS dibatasi, HTTPS aktif
- [ ] Rate limit login
- [x] Galat penyimpanan ditangani dengan rollback
- [x] Validasi alasan KEMBALIKAN di backend
- [ ] Log audit akses (siapa membuka apa), bukan hanya transisi
- [ ] Backup PostgreSQL terjadwal & pernah diuji restore
