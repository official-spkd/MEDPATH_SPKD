# Referensi API

> Semua endpoint `/api/v1`, siapa yang boleh memanggil, bentuk request & respons, dan kode status yang mungkin keluar.

## Konvensi umum

- Basis URL: `http://<host>:8080/api/v1` (dev lewat proxy: `http://localhost:5173/api/v1`).
- Body & respons: JSON UTF-8. Semua respons membawa `Cache-Control: no-store`.
- Autentikasi: `Authorization: Bearer <token>` dari `POST /auth/login`.
- Galat selalu berbentuk `{"message": "…"}`.

| Kode | Kapan |
|---|---|
| 200 | sukses |
| 204 | preflight CORS `OPTIONS` |
| 400 | JSON body tidak valid |
| 401 | token tidak ada / sesi berakhir / login gagal |
| 403 | peran tidak diizinkan (mis. `/panduan`) |
| 404 | CP tidak ditemukan (`GET /cp/{kode}`) |
| 422 | aturan domain menolak (transisi, CP terkunci, padanan tak berwenang, evaluasi CP non-aktif) |

## Ringkasan endpoint

| Metode | Jalur | Akses | Fungsi |
|---|---|---|---|
| GET | `/kesehatan` | publik | health check |
| POST | `/auth/login` | publik | masuk, dapat token |
| POST | `/auth/logout` | login | hapus token |
| GET | `/auth/me` | login | profil user aktif |
| GET | `/dashboard` | login | ringkasan eksekutif |
| GET | `/cp` | login | daftar CP |
| GET | `/cp/{kode}` | login | detail CP (+ aksi & izin ubah untuk user ini) |
| GET | `/cp/{kode}/riwayat` | login | riwayat pengesahan, terbaru dulu |
| POST | `/cp/{kode}/transisi` | login + peran sesuai aturan | AJUKAN / SETUJUI / SAHKAN / KEMBALIKAN |
| POST | `/cp/{kode}/acuan` | login + `pengubahIsi` | tambah acuan |
| DELETE | `/cp/{kode}/acuan/{id}` | idem | hapus acuan |
| POST | `/cp/{kode}/kriteria` | idem | tambah kriteria |
| DELETE | `/cp/{kode}/kriteria/{id}` | idem | hapus kriteria |
| POST | `/cp/{kode}/rencana` | idem (+ APOTEKER di Draft/Revisi) | tambah rencana |
| DELETE | `/cp/{kode}/rencana/{id}` | idem | hapus rencana |
| GET | `/approval` | login | antrean pengesahan |
| GET | `/dokumen-panduan` | login | dokumen PPK/PNPK |
| GET | `/padanan/kptl` | login | meja kerja KPTL |
| POST | `/padanan/kptl` | Tim CP/Koordinator/admin | tetapkan padanan KPTL |
| DELETE | `/padanan/kptl/{icd9cm}` | idem | cabut padanan |
| GET | `/padanan/kptl/{icd9cm}/usulan` | login | usulan KPTL otomatis |
| GET | `/padanan/snomed` | login | meja kerja SNOMED-CT |
| POST | `/padanan/snomed` | Tim CP/Koordinator/admin | tetapkan padanan SNOMED |
| DELETE | `/padanan/snomed/{icd10}` | idem | cabut padanan |
| GET | `/padanan/snomed/{icd10}/usulan` | login | usulan SNOMED otomatis |
| GET | `/evaluasi` | login | daftar evaluasi CP aktif |
| GET | `/evaluasi/{kode}` | login | evaluasi rinci satu CP aktif |
| GET | `/panduan` | **SUPER_ADMIN** | panduan ini |

## Autentikasi

### `POST /auth/login`

```json
// request
{ "email": "direktur@demo.local", "password": "Demo-SPKD-2026" }

// 200
{
  "token": "9f2c…48 karakter hex…",
  "user": { "id": 7, "nama": "Direktur (demo)", "email": "direktur@demo.local",
            "peran": "DIREKTUR", "peran_label": "Direktur" }
}
```

Email dinormalisasi (trim + huruf kecil). Email salah dan sandi salah menghasilkan pesan yang **sama** (`401 "Email atau password salah."`) agar tidak membocorkan email mana yang terdaftar.

### `GET /auth/me` → `{"user": {…}}` · `POST /auth/logout` → `{"message": "Berhasil keluar."}`

## Clinical Pathway

### `GET /cp`

```json
{ "data": [ {
  "kode": "I-4-09", "nama": "Aritmia & Gangguan Konduksi", "cmg": "I",
  "kelompok_diagnosis": "Kardiovaskular", "status": "MENUNGGU_DIREKTUR",
  "status_label": "Menunggu Direktur", "status_sejak": "2026-10-02",
  "jumlah_episode": 134, "jumlah_acuan": 2, "jumlah_rencana": 28,
  "los_median": 4, "menunggu_peran": "Direktur"
} ] }
```

### `GET /cp/{kode}`

Semua field `GET /cp` ditambah:

```json
{
  "severity":  [ { "severity": "I", "episode": 67, "los_median": 4 } ],
  "sub_cp":    [ { "icd10": "I49.9", "diagnosa": "Aritmia jantung", "jumlah_episode": 134, "porsi": "50%", "ada_acuan": true } ],
  "acuan":     [ { "id": 1, "dokumen": "PNPK …", "sumber": "PNPK", "tahun": 2023, "catatan": "…" } ],
  "kriteria":  [ { "id": 1, "jenis": "INKLUSI", "uraian": "…" } ],
  "rencana":   [ { "id": 3, "jenis": "OBAT", "severity": "I", "hari": 1, "sifat": "WAJIB",
                   "kode": "KFA-0101", "nama": "Infus kristaloid NaCl 0,9%", "dosis": "500 mL",
                   "rute": "IV", "frekuensi": "/8 jam" } ],
  "syarat":    [ { "kode": "ACUAN_CP", "judul": "…", "menghambat": true, "terpenuhi": true,
                   "penanggung_jawab": "Tim CP · KSM", "kekurangan": [] } ],
  "aksi_tersedia":       [ { "aksi": "SAHKAN", "ke": "AKTIF", "label": "Sahkan jadi Aktif" } ],
  "menunggu_peran_list": [ "Direktur" ],
  "boleh_ubah":          { "acuan": false, "rencana": false, "kriteria": false },
  "jumlah_riwayat": 3
}
```

`aksi_tersedia` dan `boleh_ubah` **bergantung pada peran pemanggil** — dua user bisa mendapat jawaban berbeda untuk CP yang sama. Semua array dijamin `[]`, tidak pernah `null`.

### `GET /cp/{kode}/riwayat`

```json
{ "data": [ { "id": 5, "dari_status": "MENUNGGU_DIREKTUR", "ke_status": "AKTIF", "aksi": "SAHKAN",
              "peran": "Direktur", "user": "Direktur (demo)", "alasan": "", "waktu": "2026-10-07 11:47" } ] }
```

### `POST /cp/{kode}/transisi`

```json
// request
{ "aksi": "KEMBALIKAN", "alasan": "Rencana severity III belum lengkap." }
// 200
{ "cp": { …CP setelah transisi… }, "message": "Status CP diperbarui." }
// 422 contoh
{ "message": "Belum bisa disahkan: Acuan standar CP — minimal satu dokumen belum terpenuhi." }
```

### Edit isi klinis

Semua endpoint edit isi membalas **detail CP lengkap terbaru** (bentuk sama dengan `GET /cp/{kode}`).

```json
POST /cp/J-4-09/acuan     { "dokumen_id": 10, "catatan": "Acuan utama PPOK" }
POST /cp/J-4-09/kriteria  { "jenis": "INKLUSI", "uraian": "PPOK eksaserbasi akut dengan sesak" }
POST /cp/J-4-09/rencana   { "jenis": "OBAT", "severity": "II", "hari": 1, "sifat": "WAJIB",
                            "kode": "KFA-0455", "nama": "Seftriakson", "dosis": "2 g",
                            "rute": "IV", "frekuensi": "/24 jam" }
DELETE /cp/J-4-09/rencana/31
```

Validasi: `jenis` kriteria harus `INKLUSI`/`EKSKLUSI`; rencana wajib `nama` dan `severity` ∈ {I, II, III}; `sifat` selain `WAJIB`/`KONDISIONAL` dijadikan `WAJIB`; `dosis`/`rute` kosong menjadi `—`.

## Antrean & referensi

- `GET /approval` → `{"data": [CP…]}` (status diproses atau REVISI).
- `GET /dokumen-panduan` → `{"data": [{ "id", "nama", "sumber", "penerbit", "tahun", "icd10_cakupan": [], "butir_count" }]}`.

## Padanan kode

### `GET /padanan/kptl` (SNOMED serupa)

```json
{
  "ringkasan": { "jumlah_prosedur": 12, "sudah": 5, "belum": 7, "cakupan_episode": 70 },
  "boleh_ubah": true,
  "master": [ { "kode": "KPTL-05", "deskripsi": "Apendektomi laparoskopik/terbuka" } ],
  "data": [ { "kode": "47.09", "deskripsi": "Apendektomi", "episode": 124, "jumlah_cp": 1,
              "padanan": "", "padanan_deskripsi": "", "status": "BELUM",
              "ditetapkan_oleh": "", "ditetapkan_pada": "" } ]
}
```

Ringkasan SNOMED memakai `jumlah_diagnosis`, `sudah`, `ragu`, `belum`. `status` ∈ `SUDAH | BELUM | RAGU`.

- `POST /padanan/kptl` `{ "icd9cm": "47.09", "kptl": "KPTL-05" }` → respons daftar lengkap terbaru.
- `POST /padanan/snomed` `{ "icd10": "J18.9", "snomed": "233604007" }`. Konsep SNOMED berstatus pensiun ditolak.
- `GET …/usulan` → `{"data": [{ "kode", "deskripsi" }]}` maksimal 5, diurutkan kemiripan.

## Evaluasi

- `GET /evaluasi` → `{"ringkasan": { "cp_aktif", "bisa_dievaluasi", "perlu_ditinjau", "belum_aktif", "klaim_sampai" }, "data": [ { "kode", "nama", "kelompok_diagnosis", "aktif_sejak", "episode_sesudah", "perlu_ditinjau", "temuan": [] } ]}`
- `GET /evaluasi/{kode}` → baris di atas + `kendali_mutu[]` (per severity) + `kendali_biaya` (bauran severity sebelum/sesudah, `tarif_rata_rata`, `total_klaim`, `persen_diagnosis_di_luar_pola`). CP yang tidak AKTIF → `422`.

## Panduan

`GET /panduan` → `{"data": [ { "id": "ikhtisar", "urutan": 1, "judul": "Ikhtisar Sistem", "ringkasan": "…", "isi": "…markdown…" } ]}`. Selain SUPER_ADMIN → `403 "Akses ini khusus Super Admin."`.

## Mencoba dengan curl

```bash
TOK=$(curl -s -X POST localhost:8080/api/v1/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"email":"super_admin@demo.local","password":"Demo-SPKD-2026"}' | jq -r .token)

curl -s localhost:8080/api/v1/cp/J-4-09 -H "Authorization: Bearer $TOK" | jq '.syarat[] | {kode, terpenuhi}'
curl -s -X POST localhost:8080/api/v1/cp/J-4-09/transisi -H "Authorization: Bearer $TOK" \
  -H 'Content-Type: application/json' -d '{"aksi":"AJUKAN"}' | jq .
```
