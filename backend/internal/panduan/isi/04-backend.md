# Backend (Go)

> Bagaimana server dirakit dari `main.go`, cara router & middleware bekerja, peran tiap file di paket `store`, dan konvensi yang wajib diikuti saat menambah kode.

## Paket & tanggung jawab

| Paket | File | Tanggung jawab |
|---|---|---|
| `main` | `main.go` | baca env, konek PostgreSQL, rakit store + router, jalankan `http.Server` |
| `internal/api` | `api.go` | daftar rute, handler, helper JSON |
| | `middleware.go` | `auth` (Bearer), `cors`, `hanyaSuperAdmin` |
| `internal/store` | `store.go` | struct `Store`, akun & token sesi, konstruktor |
| | `cpg.go` | data CP & dokumen, `ListCP`, `cpByKode` |
| | `dashboard.go` | agregasi Dashboard, antrean Approval |
| | `detail.go` | tipe isi klinis + generator isi demo |
| | `pengesahan.go` | **aturan transisi**, izin ubah isi, **7 syarat**, `Detail`, `Transisi`, riwayat |
| | `isi.go` | tambah/hapus acuan, kriteria, rencana |
| | `referensi.go` | padanan KPTL & SNOMED + usulan otomatis |
| | `evaluasi.go` | evaluasi CP aktif |
| | `persist.go` | skema PostgreSQL, seed, load, write-through |
| `internal/panduan` | `panduan.go` + `isi/*.md` | panduan ini |

Prinsip pembagian: **`api` tidak berisi logika domain**. Handler hanya membaca request, memanggil method `store`, lalu menulis JSON. Semua aturan bisnis ada di `store`, sehingga bisa diuji tanpa HTTP.

## Urutan start (`main.go`)

```go
st := store.NewDenganDB(sambungDB())   // 1. konek DB (boleh nil) → store
srv := api.New(st)                     // 2. daftarkan rute
s := &http.Server{
    Addr: alamat, Handler: srv.Handler(),           // 3. Handler = cors(mux)
    ReadTimeout: 10 * time.Second, WriteTimeout: 15 * time.Second, IdleTimeout: 60 * time.Second,
}
s.ListenAndServe()
```

`sambungDB()` membuat `pgxpool` lalu `Ping` dengan batas waktu 5 detik. Bila gagal, ia mengembalikan `nil` dan server lanjut **in-memory**. `NewDenganDB` pun begitu: bila migrasi/seed/load gagal, ia mencatat log dan memakai data memori.

## Router (`api.go`)

Memakai **ServeMux bawaan Go 1.22+** yang mengerti metode dan parameter di pola:

```go
s.mux.Handle("GET /api/v1/cp/{kode}", s.auth(s.detailCP))
// di handler:
kode := r.PathValue("kode")
```

Rute publik memakai `HandleFunc` langsung; rute terproteksi dibungkus `s.auth(...)`. Rute khusus Super Admin dibungkus dua lapis: `s.auth(s.hanyaSuperAdmin(...))`.

### Helper respons

| Helper | Perilaku |
|---|---|
| `tulisJSON(w, status, v)` | set `Content-Type: application/json; charset=utf-8` **dan** `Cache-Control: no-store`, lalu encode |
| `tulisGalat(w, status, pesan)` | `{"message": pesan}` |
| `idParam(r)` | `{id}` → `int` (0 bila bukan angka) |
| `tokenBaru()` | 24 byte acak → hex (48 karakter) |

Konvensi kode status: `400` body tidak valid, `401` token hilang/tak dikenal, `403` peran tidak boleh, `404` tidak ditemukan, `422` aturan domain menolak (mis. CP terkunci, syarat belum lengkap).

## Middleware (`middleware.go`)

**`auth`** — membaca `Authorization: Bearer <token>`, memanggil `store.UserDariToken`, menaruh `*store.User` di `context`. Handler mengambilnya dengan `userDari(r)`.

**`hanyaSuperAdmin`** — dipasang di dalam `auth`; menolak dengan `403` bila `userDari(r).Peran != "SUPER_ADMIN"`.

**`cors`** — membuka akses lintas origin (`Access-Control-Allow-Origin: *`) dan menjawab preflight `OPTIONS` dengan `204`. Di produksi sebaiknya dipersempit ke origin frontend (lihat bab Keamanan).

## Store: model konkurensi

`Store` memegang seluruh *working set* di memori dan dilindungi **satu `sync.RWMutex`**:

- method baca (`ListCP`, `Detail`, `Dashboard`, `Evaluasi`, …) → `s.mu.RLock()`
- method tulis (`Transisi`, `TambahAcuan`, `SimpanKptl`, …) → `s.mu.Lock()`

Aturan penting:

1. **Jangan memanggil method publik dari method publik lain yang sudah memegang lock** — RWMutex Go tidak re-entrant, bisa *deadlock*. Karena itu ada helper internal tanpa lock seperti `cpByKode` (komentarnya: "butuh lock dari pemanggil").
2. Data yang dikembalikan ke handler adalah **salinan** (`hasil := *cp`), bukan pointer ke state, agar tidak bisa dimutasi di luar lock.
3. Penulisan ke PostgreSQL (`simpanCPState`, `simpanRiwayat`, …) dipanggil **di dalam** lock yang sama, sehingga urutan perubahan di memori dan di DB konsisten.

## Data statis vs data yang berubah

| Jenis | Contoh | Sumber kebenaran |
|---|---|---|
| Statis (digenerasi ulang tiap start) | statistik severity, sub-CP, daftar dokumen, master KPTL/SNOMED, prosedur & diagnosis klaim | kode Go |
| Berubah saat dipakai | status CP, acuan, kriteria, rencana, riwayat, padanan | **PostgreSQL** (dimuat ke memori saat start) |

Detail mekanismenya di bab *Database & Persistensi*.

## Konvensi wajib

- **Slice kosong, bukan `nil`.** Go meng-encode `nil` slice menjadi JSON `null`. Frontend memanggil `.length` → *crash render* (halaman blank). Selalu inisialisasi `[]T{}`; lihat `bangunDetail`, `aksiUntuk`, `peranDitunggu`, `nz()`.
- **Pesan galat dalam Bahasa Indonesia** dan ramah pengguna — langsung ditampilkan di UI lewat `GalatApi.message`.
- **Nama domain dalam Bahasa Indonesia** (`Transisi`, `TambahAcuan`, `periksaSyarat`) supaya konsisten dengan istilah klinis.
- **`gofmt` sebelum commit**, dan `go vet ./...` harus bersih.
- Respons mutasi isi CP mengembalikan **detail CP lengkap** terbaru (lihat `hasilUbah`), bukan sekadar `{ok:true}`.

## Uji otomatis

```bash
cd backend && go test ./...
```

Saat ini tersedia uji untuk parser panduan dan untuk pembatasan akses `/panduan` (Super Admin `200`, peran lain `403`, tanpa token `401`). Pola uji API memakai `httptest.NewServer(api.New(store.New()).Handler())` sehingga tidak butuh PostgreSQL.
