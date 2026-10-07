# MedPath — SPKD

Clinical Pathway Generator PT SPKD. UI/UX mengikuti gaya **MedClaim** (layout), warna salmon/merah,
dengan stack **Vue (frontend) + Go (backend)**.

```
transcpg/
├── frontend/   Vue 3 + Vite (SPA)        → lihat frontend/README.md
├── backend/    Go (net/http, data demo)  → lihat backend/README.md
└── _legacy/    versi lama (arsip referensi)
    ├── backend-laravel/   backend Laravel/PHP sebelumnya
    └── frontend-nuxt/     frontend Nuxt sebelumnya
```

## Jalankan cepat

Dua terminal:

```bash
# Terminal 1 — backend (butuh PostgreSQL berjalan; lihat backend/README.md)
cd backend && go run .            # :8080

# Terminal 2 — frontend
cd frontend && npm install && npm run dev   # :5173
```

Buka http://localhost:5173 → klik salah satu chip akun demo (sandi `Demo-SPKD-2026`).

## Status

- ✅ Design system MedClaim (token warna/font, light + dark mode)
- ✅ Login (split-panel), layout app (sidebar + topbar), Dashboard Eksekutif
- ✅ CP Library, Approval, CP Aktif, Dokumen Panduan (data demo dari API Go)
- ✅ Detail CP (tab Ringkasan/Acuan/Rencana/Kriteria/Riwayat) + alur pengesahan 4 tahap
  (AJUKAN→SETUJUI→SETUJUI→SAHKAN) dengan role-gating, 7 syarat kelengkapan, dan audit trail
- ✅ Edit isi klinis (tambah/hapus acuan, kriteria, rencana) dengan penguncian per status & peran;
  syarat kelengkapan dihitung ulang langsung
- ✅ Padanan KPTL & SNOMED-CT (meja kerja + usulan otomatis) dan Evaluasi CP (kendali mutu & biaya)
- ✅ Tema salmon/merah (gaya MedClaim, bukan warnanya) — light + dark
- ✅ **Panduan Lengkap Super Admin** (Pengaturan → kartu ke-4): 13 bab dokumentasi teknis,
  ditulis di `backend/internal/panduan/isi/*.md`, hanya bisa diakses SUPER_ADMIN
- ✅ **PostgreSQL** menyimpan state (status CP, isi klinis, riwayat, padanan) — bertahan lintas restart;
  fallback in-memory bila DB mati

Butuh PostgreSQL (`brew install postgresql@17 && brew services start postgresql@17 && createdb transcpg`).
Lihat `backend/README.md`.

`_legacy/` berisi implementasi lama (Laravel + Nuxt) sebagai sumber porting logika domain CPG.
