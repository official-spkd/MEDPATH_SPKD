# Memulai: Instalasi & Menjalankan

> Prasyarat, perintah menjalankan backend dan frontend, variabel lingkungan, akun demo, dan cara mengulang data demo dari nol.

## Prasyarat

| Alat | Versi minimum | Cek | Pasang (macOS) |
|---|---|---|---|
| Go | 1.22 (dipakai 1.27) | `go version` | `brew install go` |
| Node.js + npm | Node 20+ | `node -v` | `brew install node` |
| PostgreSQL | 17 | `pg_isready` | `brew install postgresql@17` |
| Git | — | `git --version` | bawaan Xcode CLT |

Binary PostgreSQL dari Homebrew ada di `/opt/homebrew/opt/postgresql@17/bin` — tambahkan ke `PATH` bila `psql` tidak dikenali.

## 1. Siapkan database (sekali saja)

```bash
brew services start postgresql@17
createdb transcpg
```

Tidak perlu membuat tabel manual. Saat backend pertama kali konek ke database kosong, skema dibuat otomatis lalu data demo **ditanam** (seed). Jalankan berikutnya, backend **memuat** state yang tersimpan.

## 2. Jalankan backend

```bash
cd backend
go run .
```

Log yang sehat:

```text
Terhubung ke PostgreSQL.
DB kosong — menanam data demo awal ke PostgreSQL…      ← hanya saat pertama
Memuat state dari PostgreSQL (24 CP)…                  ← jalankan berikutnya
MedPath API berjalan di http://127.0.0.1:8080/api/v1
```

Kalau PostgreSQL mati atau tidak terjangkau, backend **tetap jalan** dengan data in-memory dan mencetak peringatan `… memakai data in-memory`. Semua perubahan pada mode ini hilang saat restart.

## 3. Jalankan frontend

```bash
cd frontend
npm install        # sekali, atau setelah package.json berubah
npm run dev        # http://localhost:5173
```

Vite mem-proxy setiap permintaan `/api/**` ke backend, jadi frontend dan backend tampak satu origin — tidak ada masalah CORS saat pengembangan.

## Variabel lingkungan

| Variabel | Dipakai oleh | Bawaan | Fungsi |
|---|---|---|---|
| `ALAMAT` | backend | `:8080` | alamat listen HTTP |
| `DATABASE_URL` | backend | `postgres://<user-OS>@localhost:5432/transcpg?sslmode=disable` | koneksi PostgreSQL |
| `BACKEND_URL` | Vite dev server | `http://127.0.0.1:8080` | target proxy `/api` |

Contoh:

```bash
ALAMAT=:9090 DATABASE_URL="postgres://medpath:rahasia@db:5432/transcpg?sslmode=disable" go run .
BACKEND_URL=http://127.0.0.1:9090 npm run dev
```

## Akun demo

Semua akun memakai sandi **`Demo-SPKD-2026`**. Di halaman login, klik chip peran untuk masuk instan.

| Email | Peran | Bisa apa (ringkas) |
|---|---|---|
| `super_admin@demo.local` | SUPER_ADMIN | semua aksi, Pengaturan, **Panduan Super Admin** |
| `koordinator_cp@demo.local` | KOORDINATOR_CP | menyusun, ajukan, review Tim CP, padanan kode |
| `tim_cp@demo.local` | TIM_CP | menyusun, ajukan, review Tim CP, padanan kode |
| `dpjp@demo.local` | DPJP | menyusun isi & mengajukan CP Draft/Revisi |
| `apoteker@demo.local` | APOTEKER | mengubah **rencana** pada CP Draft/Revisi |
| `komite_medik@demo.local` | KOMITE_MEDIK | menyetujui/mengembalikan di Review Komite, mengembalikan CP Aktif |
| `direktur@demo.local` | DIREKTUR | mengesahkan/mengembalikan di Menunggu Direktur |

Matriks lengkapnya ada di bab *Alur Pengesahan & Aturan Domain*.

## Mengulang data demo dari nol

```bash
psql -d transcpg -c "DROP TABLE IF EXISTS cp_state,cp_acuan,cp_kriteria,cp_rencana,cp_riwayat,padanan_kptl,padanan_snomed CASCADE;"
cd backend && go run .     # skema dibuat ulang + seed
```

Setelah backend restart, **semua sesi login hilang** (token disimpan di memori) — pengguna perlu login ulang.

## Perintah harian

| Tujuan | Perintah |
|---|---|
| Format kode Go | `cd backend && gofmt -w .` |
| Cek statis Go | `go vet ./...` |
| Uji Go | `go test ./...` |
| Build binary backend | `go build -o transcpg-api .` |
| Build frontend produksi | `cd frontend && npm run build` → `frontend/dist/` |
| Pratinjau build | `npm run preview` |
| Siapa yang memakai port | `lsof -ti:8080` / `lsof -ti:5173` |

Catatan: `npm run build` hanya menjalankan `vite build` (tidak ada pemeriksaan tipe `vue-tsc` di skrip build). Untuk cek tipe manual: `npx vue-tsc --noEmit`.
