# Ikhtisar Sistem

> Apa itu MedPath, bagaimana bagian-bagiannya tersusun, dan ke mana sebuah klik di layar mengalir sampai ke database.

## Apa itu MedPath

**MedPath** adalah *Clinical Pathway Generator* milik PT SPKD. Aplikasi ini membantu rumah sakit menyusun **Clinical Pathway (CP)** dari pola klaim INA-CBG, melengkapinya dengan acuan klinis (PNPK/PPK), rencana isi klinis per severity, serta kriteria inklusi/eksklusi, lalu mengesahkannya berjenjang dari Tim CP sampai Direktur. CP yang sudah **AKTIF** kemudian dievaluasi terhadap klaim sesudahnya.

Nama folder & modul Go adalah `transcpg` (historis). **Nama produk yang tampil ke pengguna selalu "MedPath".** Tampilan mengikuti tata letak produk saudara *MedClaim*, tetapi dengan palet **salmon/merah**.

## Stack teknologi

| Lapisan | Teknologi | Versi | Lokasi |
|---|---|---|---|
| Frontend | Vue 3 (`<script setup>` + TypeScript), Vite | Vue 3.5 · Vite 6 | `frontend/` |
| Routing | vue-router, **hash history** (`#/…`) | 4.x | `frontend/src/router` |
| State | Pinia | 2.x | `frontend/src/stores` |
| Ikon | lucide-vue-next | 0.460 | — |
| Grafik | SVG buatan sendiri (Gauge, Donut, Area) | — | `frontend/src/components` |
| Backend | Go, `net/http` (ServeMux Go 1.22+) | Go 1.27 | `backend/` |
| Driver DB | pgx v5 (`pgxpool`) | 5.11 | `backend/internal/store/persist.go` |
| Database | PostgreSQL | 17 | db `transcpg` |
| Markdown panduan | `go:embed` + `marked` di frontend | — | `backend/internal/panduan` |

Tidak ada framework CSS (Tailwind dsb.) dan tidak ada framework HTTP (Gin/Echo). Semua sengaja ramping.

## Struktur repositori

```text
transcpg/                      ← repo: github.com/official-spkd/MEDPATH_SPKD (private)
├── README.md
├── .gitignore                 ← mengecualikan node_modules, dist, .env, _legacy/
├── frontend/                  ← SPA Vue 3 + Vite
│   ├── index.html             ← font, favicon, skrip tema anti-kedip
│   ├── vite.config.ts         ← alias @ → src, proxy /api → backend
│   └── src/
│       ├── main.ts            ← bootstrap app + errorHandler global
│       ├── App.vue
│       ├── router/index.ts    ← daftar rute + guard login/admin/super admin
│       ├── stores/            ← auth.ts (sesi), tema.ts (light/dark)
│       ├── lib/               ← api.ts (klien HTTP), format.ts (angka & label)
│       ├── layouts/AppLayout.vue
│       ├── components/        ← Sidebar, Topbar, StatCard, StatusPill, grafik, dll.
│       ├── pages/             ← satu file per halaman
│       └── assets/            ← tokens.css (design token), base.css (util)
├── backend/                   ← API Go
│   ├── main.go                ← konek DB, rakit store & router, jalankan server
│   └── internal/
│       ├── api/               ← router, middleware auth/CORS, handler
│       ├── store/             ← seluruh logika domain + persistensi
│       └── panduan/           ← panduan ini (Markdown ter-embed)
└── _legacy/                   ← arsip Laravel + Nuxt lama (TIDAK ikut ke git)
```

## Arsitektur (level container)

```text
 Pengguna (10 peran dikenali backend)
        │
        ▼
 ┌─────────────── Browser ───────────────┐
 │  SPA Vue 3 (hash router, Pinia)        │
 │  lib/api.ts  ── fetch + Bearer token ──┼──┐
 └────────────────────────────────────────┘  │  dev: lewat proxy Vite :5173
                                              ▼
 ┌──────────── Server Go :8080 ────────────────────────────┐
 │ ServeMux /api/v1 → cors → auth(Bearer) → handler        │
 │                                  │                      │
 │                                  ▼                      │
 │   store: aturan pengesahan · syarat · evaluasi · padanan│
 │          (working set di memori, dikunci RWMutex)       │
 │                                  │ write-through        │
 └──────────────────────────────────┼──────────────────────┘
                                    ▼
                         PostgreSQL 17 (7 tabel state)
```

Diagram C4 lengkap dalam format draw.io juga tersedia: `MedPath_Arsitektur_C4.drawio`.

## Perjalanan satu permintaan

Contoh: Direktur menekan **"Sahkan jadi Aktif"** di halaman Detail CP.

1. `CpDetail.vue` memanggil `api('cp/I-4-09/transisi', { method: 'POST', body: '{"aksi":"SAHKAN"}' })`.
2. `lib/api.ts` menambahkan header `Authorization: Bearer <token>` (token dari `localStorage`).
3. Di mode dev, Vite mem-proxy `/api/**` ke `http://127.0.0.1:8080`.
4. Go: `cors()` → ServeMux mencocokkan `POST /api/v1/cp/{kode}/transisi` → `auth()` membaca token, mencari user, menaruhnya di `context`.
5. Handler `transisiCP` memanggil `store.Transisi(kode, aksi, user, alasan)`.
6. `Transisi` mengunci store, memastikan aksi tersedia di status sekarang, peran user berwenang, dan — karena tujuannya AKTIF — **syarat menghambat** terpenuhi.
7. Status berubah di memori, riwayat ditambahkan, lalu **write-through** ke tabel `cp_state` dan `cp_riwayat`.
8. Handler membalas JSON; frontend memuat ulang detail dan memperbarui badge antrean Approval di sidebar.

## Peta halaman

| Menu | Rute | Halaman | Endpoint utama |
|---|---|---|---|
| Login | `#/login` | `Login.vue` | `POST /auth/login` |
| Dashboard | `#/` | `Dashboard.vue` | `GET /dashboard` |
| CP Library | `#/library` | `CpLibrary.vue` | `GET /cp` |
| Detail CP | `#/library/:kode` | `CpDetail.vue` | `GET /cp/{kode}`, `/riwayat`, `/transisi`, edit isi |
| Approval | `#/approval` | `Approval.vue` | `GET /approval` |
| CP Aktif | `#/cp-aktif` | `CpAktif.vue` | `GET /cp` (disaring AKTIF) |
| Evaluasi CP | `#/evaluasi` | `Evaluasi.vue` | `GET /evaluasi`, `/evaluasi/{kode}` |
| Dokumen Panduan | `#/dokumen` | `Dokumen.vue` | `GET /dokumen-panduan` |
| Padanan KPTL / SNOMED | `#/padanan/kptl`, `#/padanan/snomed` | `Padanan.vue` | `/padanan/kptl`, `/padanan/snomed` |
| Pengaturan | `#/pengaturan` | `Pengaturan.vue` | — (admin) |
| Panduan Super Admin | `#/pengaturan/panduan` | `PanduanSuperAdmin.vue` | `GET /panduan` (SUPER_ADMIN) |

## Status kematangan (jujur)

- **Sudah nyata:** alur pengesahan, aturan peran, 7 syarat, edit isi klinis, padanan, persistensi PostgreSQL.
- **Masih data demo/sintetis:** daftar CP & statistiknya, dokumen panduan, master kode, tren dashboard, angka evaluasi.
- **Belum ada:** impor klaim INA-CBG, pembangun library otomatis, API integrasi TransCPR-X, hashing password, token sesi persisten. Detailnya di bab *Keamanan & RBAC* dan *Deploy & Operasional*.
