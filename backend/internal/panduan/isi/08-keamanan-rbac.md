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

1. `POST /auth/login` mencocokkan email + sandi dengan daftar user di memori.
2. Bila cocok, server membuat token acak 24 byte (`crypto/rand`, hex) dan menyimpannya di peta `tokens map[string]int` (token → id user).
3. Frontend menyimpan token di `localStorage` dan mengirimnya sebagai `Bearer` di setiap request.
4. `POST /auth/logout` menghapus token dari peta.

## Keterbatasan saat ini (WAJIB dibaca)

| # | Kondisi sekarang | Risiko | Perbaikan |
|---|---|---|---|
| 1 | Sandi akun disimpan **polos** di kode (`Demo-SPKD-2026`) | siapa pun yang membaca repo tahu sandinya | tabel `users` + `bcrypt`/`argon2id`; hapus sandi demo dari build produksi |
| 2 | Token sesi hanya di **memori**, tanpa kedaluwarsa | sesi hilang tiap restart; token yang bocor berlaku selamanya selama server hidup | tabel `sesi` dengan `expires_at`, atau JWT berumur pendek + refresh |
| 3 | Token di `localStorage` | rentan dicuri jika ada XSS | pertimbangkan cookie `HttpOnly; Secure; SameSite=Strict` |
| 4 | CORS `Access-Control-Allow-Origin: *` | situs lain bisa memanggil API (meski tetap butuh token) | batasi ke origin frontend produksi |
| 5 | Tidak ada pembatasan laju login | tebak-sandi tanpa batas | rate limit per IP/email, kunci sementara |
| 6 | Tidak ada HTTPS di server Go | token lewat jalur polos jika diekspos langsung | pasang di belakang reverse proxy TLS (Caddy/Nginx) |
| 7 | Alasan KEMBALIKAN hanya divalidasi di UI | klien lain bisa mengembalikan tanpa alasan | validasi di `Transisi` |
| 8 | Write-through mengabaikan galat DB | audit trail bisa tidak tersimpan tanpa ketahuan | kembalikan galat & batalkan perubahan memori |

## Data pasien

Seluruh data di MedPath saat ini **fiktif dan tanpa identitas pasien**. CP bekerja di level agregat (episode per kode INA-CBG/ICD), bukan rekam medis individual. Saat impor klaim riil dibangun, pastikan hanya field agregat yang dimuat (kode grouper, severity, LOS, kelas, tarif, diagnosis) — **tanpa** nama, NIK, nomor RM, atau nomor SEP yang bisa ditelusuri.

## Rahasia & repositori

- `.env` dan file `*.local` di-*ignore* git.
- Repo `official-spkd/MEDPATH_SPKD` bersifat **private**.
- `gh` CLI di mesin pengembang tidak login sebagai official-spkd; push memakai Personal Access Token milik official-spkd yang ditempel sendiri oleh pemilik akun saat diminta git. Cabut token yang tidak dipakai lagi di *Settings → Developer settings → Personal access tokens*.

## Daftar periksa sebelum produksi

- [ ] Sandi di-hash, sandi demo dihapus
- [ ] Sesi persisten dengan kedaluwarsa
- [ ] CORS dibatasi, HTTPS aktif
- [ ] Rate limit login
- [ ] Galat write-through ditangani
- [ ] Validasi alasan KEMBALIKAN di backend
- [ ] Log audit akses (siapa membuka apa), bukan hanya transisi
- [ ] Backup PostgreSQL terjadwal & pernah diuji restore
