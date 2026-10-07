# MedPath — Backend (Go)

Clinical Pathway Generator (CPG) versi SPKD. Server API untuk frontend Vue.
Ditulis dengan **Go + `net/http`** (stdlib untuk HTTP) dan **PostgreSQL** (via `pgx`) untuk state.

> Semua data FIKTIF (tanpa identitas pasien). State yang berubah saat runtime (status CP,
> isi klinis, riwayat pengesahan, padanan) **disimpan di PostgreSQL**; data statis (statistik
> severity, dokumen, master kode) digenerasi di Go. Bila DB tidak tersedia, server otomatis
> **fallback ke in-memory** (perubahan tidak tersimpan).

## Menjalankan

### 1. PostgreSQL (sekali setup)

```bash
brew install postgresql@17
brew services start postgresql@17
createdb transcpg
```

Server memakai `DATABASE_URL`, atau default `postgres://<user>@localhost:5432/transcpg?sslmode=disable`.
Saat pertama jalan dengan DB kosong, skema dibuat otomatis dan data demo ditanam. Restart berikutnya
memuat state dari DB.

### 2. Server

```bash
cd backend
go run .                 # :8080, konek PostgreSQL
ALAMAT=:9090 go run .     # port lain
DATABASE_URL=... go run . # DB lain
```

Mulai ulang data demo dari awal:

```bash
psql -d transcpg -c "DROP TABLE IF EXISTS cp_state,cp_acuan,cp_kriteria,cp_rencana,cp_riwayat,padanan_kptl,padanan_snomed CASCADE;"
go run .   # akan menanam ulang
```

Build biner:

```bash
go build -o transcpg-api .
./transcpg-api
```

## Akun demo

Semua akun pakai sandi **`Demo-SPKD-2026`**:

| Email | Peran |
|---|---|
| `super_admin@demo.local` | Super Admin |
| `koordinator_cp@demo.local` | Koordinator CP |
| `tim_cp@demo.local` | Tim CP |
| `dpjp@demo.local` | DPJP |
| `apoteker@demo.local` | Apoteker |
| `komite_medik@demo.local` | Komite Medik |
| `direktur@demo.local` | Direktur |

## Endpoint (`/api/v1`)

| Metode | Jalur | Proteksi | Fungsi |
|---|---|---|---|
| POST | `/auth/login` | — | Masuk, kembalikan token + user |
| POST | `/auth/logout` | Bearer | Keluar |
| GET | `/auth/me` | Bearer | Profil user aktif |
| GET | `/dashboard` | Bearer | Ringkasan eksekutif (readiness, kartu, tren, dll) |
| GET | `/cp` | Bearer | Daftar seluruh Clinical Pathway |
| GET | `/cp/{kode}` | Bearer | Detail CP (severity, sub-CP, acuan, rencana, kriteria, syarat, aksi tersedia per peran) |
| GET | `/cp/{kode}/riwayat` | Bearer | Riwayat pengesahan (audit trail) |
| POST | `/cp/{kode}/transisi` | Bearer | Jalankan aksi pengesahan (`AJUKAN`/`SETUJUI`/`SAHKAN`/`KEMBALIKAN`) sesuai peran |
| POST/DELETE | `/cp/{kode}/acuan`, `/kriteria`, `/rencana` | Bearer | Edit isi klinis (terkunci per status & peran) |
| GET | `/approval` | Bearer | Antrean pengesahan |
| GET | `/dokumen-panduan` | Bearer | Daftar dokumen panduan |
| GET/POST/DELETE | `/padanan/kptl`, `/padanan/snomed` | Bearer | Meja kerja padanan kode (+`/{kode}/usulan`) |
| GET | `/evaluasi`, `/evaluasi/{kode}` | Bearer | Evaluasi CP aktif (kendali mutu & biaya) |
| GET | `/panduan` | Bearer **SUPER_ADMIN** | Panduan Lengkap Super Admin (selain itu `403`) |
| GET | `/kesehatan` | — | Health check |

Autentikasi: header `Authorization: Bearer <token>`. Token sesi disimpan di memori
(hilang saat server restart).

## Struktur

```
backend/
├── main.go                     # entry point + konfigurasi HTTP server
├── internal/
│   ├── api/
│   │   ├── api.go              # router + handler
│   │   └── middleware.go       # auth (Bearer) + CORS
│   └── store/
│       ├── store.go           # user & token sesi
│       ├── cpg.go             # data CP, dokumen (demo)
│       └── dashboard.go       # penyusun payload dashboard
└── go.mod
```

## Menuju produksi

- Ganti `store` dengan PostgreSQL (mis. `pgx`), migrasi dari skema Laravel lama di `_legacy/backend-laravel`.
- Hash password (bcrypt/argon2) — demo ini memakai plaintext.
- Token JWT ber-expiry atau sesi di DB, bukan map memori.
- Port logika domain CPG (aturan status, pengesahan 4 tahap, impor klaim, pembangun library) dari `_legacy/backend-laravel/app/Cpg`.
- Batasi `Access-Control-Allow-Origin` ke origin frontend produksi.
