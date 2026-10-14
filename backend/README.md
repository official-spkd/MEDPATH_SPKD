# MedPath — Backend (Go)

Clinical Pathway Generator (CPG) versi SPKD. Server API untuk frontend Vue.
Ditulis dengan **Go + `net/http`** (stdlib untuk HTTP) dan **PostgreSQL** (via `pgx`) untuk state.

> Semua data FIKTIF (tanpa identitas pasien). State yang berubah saat runtime (status CP,
> isi klinis, riwayat pengesahan, padanan) **disimpan di PostgreSQL**; data statis (statistik
> severity, dokumen, master kode) digenerasi di Go. Bila DB tidak tersedia, startup gagal atau API mengembalikan **503**.
> Mode memori hanya untuk lokal dengan `DEMO_IN_MEMORY=true`.

## Menjalankan

### 1. PostgreSQL (sekali setup)

```bash
brew install postgresql@17
brew services start postgresql@17
createdb transcpg
```

Server memakai `DATABASE_URL`, atau default `postgres://<user>@localhost:5432/transcpg?sslmode=disable`.
Saat pertama jalan dengan DB kosong, skema dibuat otomatis dan data demo ditanam. Restart berikutnya
memuat snapshot state dari DB per request.

### 2. Server

```bash
cd backend
go run .                 # :8080, konek PostgreSQL
ALAMAT=:9090 go run .     # port lain
DATABASE_URL=... go run . # DB lain
```

Mulai ulang data demo dari awal:

```bash
psql -d transcpg -c "DROP TABLE IF EXISTS auth_sessions,cp_state,cp_acuan,cp_kriteria,cp_rencana,cp_riwayat,padanan_kptl,padanan_snomed CASCADE;"
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

Autentikasi: header `Authorization: Bearer <token>`. Hash token sesi disimpan di tabel `auth_sessions`, kedaluwarsa setelah 24 jam,
dan logout berlaku lintas instance. Respons API memakai `Cache-Control: no-store`.

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

- Migrasi pengguna dan data nyata sesuai kebutuhan operasional; data awal masih demo.
- Hash password (bcrypt/argon2) — demo ini memakai plaintext.
- Jadwalkan pembersihan sesi kedaluwarsa jika jumlah pengguna meningkat.
- Port logika domain CPG (aturan status, pengesahan 4 tahap, impor klaim, pembangun library) dari `_legacy/backend-laravel/app/Cpg`.
- Batasi `Access-Control-Allow-Origin` ke origin frontend produksi.

## Konsistensi di Vercel

`Store.Request` membuat snapshot terpisah dalam transaksi untuk setiap permintaan.
Pembaca memakai repeatable-read. Penulis memakai advisory lock transaksi PostgreSQL,
sehingga validasi status/peran, penentuan ID, perubahan CP, dan audit tidak berlomba
antar-instance. Respons sukses dikirim setelah commit; kegagalan SQL menjadi 503.
Untuk skala data yang lebih besar, ganti pemuatan seluruh state dengan query per objek
serta lock per CP. Saat ini pendekatan ini mempertahankan seluruh logika domain demo.

Koneksi memakai pgx mode `exec` (tanpa prepared statement bernama), maksimum 3 koneksi
per instance dan idle 30 detik, sesuai transaction pooler Supabase.

Pengujian integrasi membuat lalu menghapus schema acak di database uji:

```bash
TEST_DATABASE_URL='postgres://localhost/medpath_test?sslmode=disable' go test -race ./...
```

Jangan arahkan `TEST_DATABASE_URL` ke database produksi.
