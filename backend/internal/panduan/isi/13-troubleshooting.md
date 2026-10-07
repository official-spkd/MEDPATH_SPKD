# Troubleshooting & FAQ

> Gejala yang pernah atau mungkin muncul, penyebabnya, dan cara memperbaikinya — disusun dari masalah nyata selama pengembangan MedPath.

## Frontend

| Gejala | Penyebab paling mungkin | Perbaikan |
|---|---|---|
| Halaman tiba-tiba **kosong/blank**, layout masih ada | galat saat render; sering `Cannot read properties of null (reading 'length')` | buka Console, cari `[VUE ERR]`; pastikan backend mengirim `[]` bukan `null`; pakai `?.length` untuk data opsional |
| Data tidak berubah setelah aksi berhasil | respons GET tersaji dari cache peramban | pastikan request lewat `api()` (punya cache-buster) dan backend mengirim `Cache-Control: no-store` |
| Stuck di "Memuat…" setelah memperbarui rute di dev | Vite gagal me-*resolve* import dinamis (mis. rute ditambah sebelum file halamannya ada), chunk gagal ter-cache | restart `npm run dev`, muat ulang keras halaman |
| Pindah dari CP A ke CP B tapi isi tetap CP A | komponen dipakai ulang; `onMounted` tidak terpanggil | muat data lewat `watch(param, muat, { immediate: true })` |
| Tiba-tiba dilempar ke halaman Login | backend restart → token tidak dikenal (`401`) | login ulang (perilaku normal sampai sesi persisten dibangun) |
| Menu Pengaturan / Panduan tidak muncul | peran tidak memenuhi | Pengaturan: admin; Panduan: SUPER_ADMIN |
| Warna emas/kuning muncul di komponen baru | hex ditulis langsung | ganti dengan token `var(--brand…)` |

## Backend

| Gejala | Penyebab | Perbaikan |
|---|---|---|
| `listen tcp :8080: bind: address already in use` | proses lama masih jalan | `lsof -ti:8080 \| xargs kill` lalu jalankan lagi |
| Log `… memakai data in-memory` | PostgreSQL mati / `DATABASE_URL` salah | `brew services start postgresql@17`, `pg_isready`, cek DSN |
| `duplicate key value violates unique constraint` saat seed | tabel setengah terisi dari seed gagal sebelumnya | jalankan perintah reset data demo (bab *Memulai*) |
| `422 Aksi "X" tidak tersedia pada status Y` | aksi tidak berlaku di status itu | cek matriks di bab *Alur Pengesahan* |
| `422 Peran Anda tidak berwenang…` | peran tidak termasuk aksi | login dengan peran yang benar |
| `422 Belum bisa disahkan: … belum terpenuhi` | syarat menghambat belum lengkap | lengkapi acuan / rencana tiap severity ≥ 10 episode |
| `422 Isi CP terkunci…` | CP di Review Komite / Menunggu Direktur / Aktif | KEMBALIKAN dulu ke REVISI |
| `403 Akses ini khusus Super Admin.` | bukan SUPER_ADMIN | gunakan akun Super Admin |
| `go: command not found` | Go terpasang via Homebrew tapi PATH belum memuat | `export PATH="/opt/homebrew/bin:$PATH"` |
| `psql: command not found` | biner PostgreSQL 17 tidak di PATH | `export PATH="/opt/homebrew/opt/postgresql@17/bin:$PATH"` |

## Git & GitHub

| Gejala | Penyebab | Perbaikan |
|---|---|---|
| `403` saat `git push` ke MEDPATH_SPKD | git memakai kredensial akun lain dari Keychain | `git -c credential.helper= push -u origin main`, tempel token official-spkd |
| `gh repo create` membuat repo di akun yang salah | `gh` hanya login sebagai akun lain | buat repo lewat web di sesi official-spkd, atau `gh auth login` untuk official-spkd |
| File `_legacy/`, `node_modules/`, `.env` ikut ter-commit | `.gitignore` tidak terbaca | cek `.gitignore` di root; `git rm -r --cached <path>` |

## FAQ

**Kenapa folder bernama `transcpg` padahal produk bernama MedPath?**
Nama historis. Folder, modul Go (`module transcpg`), dan nama database tetap `transcpg`; semua teks yang terlihat pengguna wajib "MedPath".

**Kenapa ada `_legacy/`?**
Berisi versi lama Laravel + Nuxt. Tidak dijalankan dan tidak ikut ke git, tetapi menjadi rujukan logika domain lengkap (impor klaim, pembangun library, evaluasi berbasis klaim) yang belum diport.

**Apakah data di aplikasi ini data pasien sungguhan?**
Tidak. Semuanya fiktif dan agregat.

**Bagaimana menambah isi panduan ini?**
Ikuti *Resep 7* di bab *Resep Pengembangan*.

**Siapa yang bisa melihat panduan ini?**
Hanya akun berperan `SUPER_ADMIN`. Pembatasannya ditegakkan di server (`403` untuk peran lain), bukan sekadar menu yang disembunyikan.
