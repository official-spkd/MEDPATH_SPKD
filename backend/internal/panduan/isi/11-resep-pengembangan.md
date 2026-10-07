# Resep Pengembangan

> Langkah konkret untuk perubahan yang paling sering: halaman baru, endpoint baru, endpoint khusus peran, data yang perlu disimpan, peran baru, syarat baru, dan bab panduan baru.

## Resep 1 — Menambah halaman baru

1. Buat `frontend/src/pages/NamaHalaman.vue` mengikuti pola standar (lihat bab *Design System*).
2. Daftarkan rute di `frontend/src/router/index.ts` sebagai anak `/`:
   ```ts
   { path: 'laporan', name: 'laporan', component: () => import('@/pages/Laporan.vue'),
     meta: { judul: 'Laporan' } },
   ```
   Tambahkan `admin: true` atau `superAdmin: true` di `meta` bila perlu dibatasi.
3. Tambahkan menu di `components/Sidebar.vue` (array `grup`, pilih grup *Utama* atau *Referensi*, ikon dari lucide).
4. Ambil data dengan `api<T>()` + tangani `GalatApi`.
5. `npm run build` harus sukses.

## Resep 2 — Menambah endpoint baru

1. **Logika di `store`**, bukan di handler:
   ```go
   // store/laporan.go
   func (s *Store) Laporan() []BarisLaporan {
       s.mu.RLock()
       defer s.mu.RUnlock()
       out := []BarisLaporan{}          // ← bukan nil!
       // …
       return out
   }
   ```
2. **Handler tipis** di `api/api.go`:
   ```go
   func (s *Server) laporan(w http.ResponseWriter, _ *http.Request) {
       tulisJSON(w, http.StatusOK, map[string]any{"data": s.store.Laporan()})
   }
   ```
3. **Daftarkan rute** di `rute()`:
   ```go
   s.mux.Handle("GET /api/v1/laporan", s.auth(s.laporan))
   ```
4. `gofmt -w . && go vet ./... && go test ./...`, lalu restart backend.

## Resep 3 — Endpoint yang hanya boleh untuk peran tertentu

Untuk Super Admin sudah ada pembungkus siap pakai:

```go
s.mux.Handle("GET /api/v1/rahasia", s.auth(s.hanyaSuperAdmin(s.rahasia)))
```

Untuk aturan yang bergantung pada **status data** (bukan sekadar peran), taruh pengecekan di `store` seperti `bolehUbah(peran, status, objek)` dan kembalikan `error` berbahasa Indonesia; handler memetakannya ke `422`.

Untuk sekelompok peran, buat pembungkus serupa di `middleware.go`:

```go
func (s *Server) hanyaPeran(daftar []string, h http.HandlerFunc) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        u := userDari(r)
        for _, p := range daftar {
            if u != nil && u.Peran == p { h(w, r); return }
        }
        tulisGalat(w, http.StatusForbidden, "Peran Anda tidak berwenang.")
    }
}
```

## Resep 4 — Data baru yang harus tersimpan

1. Tambahkan `CREATE TABLE IF NOT EXISTS …` ke konstanta `skema` di `persist.go`.
2. Tambahkan penulisan awal di `seedDB()` dan pembacaan di `loadDB()`.
3. Buat helper write-through (`simpanXDB`) yang *no-op* bila `s.pool == nil`.
4. Panggil helper itu **di dalam lock** pada method store yang memutasi data.
5. Uji siklus penuh: ubah lewat UI → restart backend → data masih ada.

Catatan: menambah kolom ke tabel yang sudah ada **tidak** terjadi otomatis (`IF NOT EXISTS` tidak mengubah tabel lama). Pakai `ALTER TABLE … ADD COLUMN IF NOT EXISTS …` di `skema`, atau reset data demo.

## Resep 5 — Menambah peran baru

1. `store/store.go` → tambahkan ke `seed` di `newBase()` bila perlu akun demo (email otomatis `<peran huruf kecil>@demo.local`).
2. `store/pengesahan.go` → tambahkan label di `labelPeran`, lalu masukkan ke kelompok yang tepat (`penyusun()`, `peranTimCp()`, atau peran di `transisiDari`/`pengubahIsi`).
3. Frontend → bila peran itu admin, tambahkan ke `ADMIN` di `stores/auth.ts`; tambahkan chip di `Login.vue` bila perlu.
4. Perbarui tabel peran di bab *Keamanan & RBAC* dan *Alur Pengesahan*.

## Resep 6 — Menambah syarat kelengkapan

Di `periksaSyarat()` (`pengesahan.go`), hitung daftar `kekurangan []string`, lalu tambahkan:

```go
syarat("KODE_BARU", "Judul yang dibaca manusia", false /* menghambat? */, "Penanggung jawab", kurangBaru),
```

Bila `menghambat = true`, syarat itu otomatis memblokir SAHKAN (loop di `Transisi`). UI menampilkannya tanpa perubahan frontend.

## Resep 7 — Menambah atau mengubah bab panduan ini

1. Buat/ubah file di `backend/internal/panduan/isi/`, nama `NN-slug.md` (`NN` menentukan urutan, `slug` menjadi id).
2. Baris pertama **wajib** `# Judul`; paragraf `> …` tepat di bawahnya menjadi ringkasan.
3. Markdown GitHub didukung: tabel, blok kode, daftar, kutipan, daftar centang.
4. `go test ./internal/panduan/` memastikan semua file valid.
5. Build ulang & restart backend (file di-*embed* ke binary).

## Resep 8 — Alur Git

```bash
cd ~/Documents/transcpg
git checkout -b fitur/nama-fitur
# … kerjakan …
cd backend && gofmt -w . && go vet ./... && go test ./... && cd ..
cd frontend && npm run build && cd ..
git add -A && git commit -m "Tambah laporan bulanan CP"
git -c credential.helper= push -u origin fitur/nama-fitur   # tempel token official-spkd
```

Buka Pull Request di `github.com/official-spkd/MEDPATH_SPKD`, gabungkan ke `main` setelah diperiksa.
