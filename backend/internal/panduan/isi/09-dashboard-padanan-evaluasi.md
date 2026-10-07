# Dashboard, Padanan & Evaluasi

> Rumus di balik angka Dashboard, cara usulan padanan kode dihitung, dan aturan yang memunculkan "temuan" di Evaluasi CP — termasuk mana yang masih data sintetis.

## Dashboard (`store/dashboard.go`)

Satu panggilan `GET /dashboard` menghitung semuanya dari daftar CP saat itu.

### Kartu KPI

| Kartu | Rumus |
|---|---|
| Clinical Pathway | jumlah CP |
| CP tanpa acuan | CP dengan `jumlah_acuan == 0` |
| Dalam pengesahan | CP berstatus `sedangDiproses` |
| CP aktif | CP berstatus AKTIF |
| Dokumen panduan | jumlah dokumen |
| Butir acuan klinis | jumlah seluruh butir per kategori |

### MedPath Readiness Index

Lima komponen, masing-masing rasio `nilai / maks` dibatasi maksimal 1:

| Komponen | nilai | maks |
|---|---|---|
| Acuan klinis ditetapkan | CP yang punya acuan | jumlah CP |
| Rencana klinis terisi | CP yang punya rencana | jumlah CP |
| CP tersahkan aktif | CP AKTIF | jumlah CP |
| Dokumen panduan terhubung | jumlah dokumen | 14 (target) |
| Butir acuan terekstrak | jumlah butir | 220 (target) |

`skor = rata-rata(rasio) × 100`, lalu dilabeli: ≥ 85 *sangat siap*, ≥ 70 *cukup siap*, ≥ 50 *perlu dilengkapi*, selain itu *masih awal*.

### Bagian lain

- **Distribusi status** → hitungan per status.
- **Kasus per kelompok diagnosis** → total episode per CMG, urut terbesar.
- **Prioritas penyusunan** → 8 CP dengan episode terbanyak.
- **Tren** → lima bulan pertama **masih angka tetap** di kode; hanya bulan terakhir dihitung dari data nyata.
- **Tindakan prioritas** → **teks tetap** di kode (belum dihitung dinamis).

## Padanan kode (`store/referensi.go`)

Tujuan: setiap prosedur ICD-9-CM di klaim punya padanan **KPTL** (agar rencana tindakan selaras tarif lokal), dan setiap diagnosis ICD-10 punya padanan **SNOMED-CT** (interoperabilitas rekam medis).

- Daftar prosedur/diagnosis, master KPTL, dan master SNOMED saat ini adalah **data demo** di kode.
- Penetapan padanan disimpan di tabel `padanan_kptl` / `padanan_snomed` beserta siapa & kapan.
- Status SNOMED `RAGU` menandai usulan otomatis yang belum dikonfirmasi manusia; RAGU **tidak dihitung** sebagai "sudah".
- **Cakupan episode** KPTL = total episode prosedur yang sudah dipetakan ÷ total episode semua prosedur.

### Algoritma usulan (`usulanKode`)

Kemiripan kata sederhana (indeks Jaccard):

1. Pecah deskripsi kode sumber menjadi kata huruf kecil; buang kata < 4 huruf.
2. Lakukan hal sama untuk tiap deskripsi di master (hanya yang aktif).
3. `skor = |kata sama| / |gabungan kata|`.
4. Ambil 5 skor tertinggi yang > 0.

Contoh: "Apendektomi" ↔ "Apendektomi laparoskopik/terbuka" → kata sama `apendektomi` → masuk usulan teratas.

> Usulan **hanya petunjuk** dan sering keliru untuk istilah yang berbeda kata tapi sama makna. Keputusan tetap di Tim Koding — itulah sebabnya penetapan harus diklik manual.

## Evaluasi CP (`store/evaluasi.go`)

Tujuan konsep: membandingkan klaim **sesudah** CP aktif dengan target yang berlaku **saat CP disahkan**, pada dua sisi — mutu (lama rawat) dan biaya (tarif INA-CBG). Biaya riil rumah sakit tidak dimuat.

> **Penting:** karena impor klaim belum ada, angka evaluasi saat ini **sintetis tapi deterministik** — dibangkitkan dari hash kode CP (`hashKode`), sehingga selalu sama untuk CP yang sama. Logika temuan-nya sudah nyata; sumber datanya yang belum.

### Kendali mutu (per severity)

- `target_los` = LOS median severity itu; `target_p75 = target_los + 2`.
- Dibandingkan dengan LOS median & persentase episode yang melewati p75 target sesudah CP aktif.

### Aturan temuan

| Aturan | Pesan |
|---|---|
| median LOS sesudah > p75 target | "Severity X: median LOS N hari melebihi p75 target (M hari)." |
| ≥ 30% episode melewati p75 target | "Severity X: P% episode lama rawatnya melewati p75 target." |
| porsi severity III naik ≥ 8 poin persentase | "Porsi severity III naik … Periksa kesesuaian pengkodean." |

CP dengan ≥ 1 temuan ditandai **perlu ditinjau**. Kenaikan porsi severity III adalah sinyal klasik *upcoding*, karena itu dijadikan temuan tersendiri.

### Kendali biaya

Tarif rata-rata, total klaim sesudah aktif, persentase diagnosis di luar pola CP, dan bauran severity sebelum → sesudah.

### Ke versi produksi

Rujuk `_legacy/backend-laravel/app/Http/Controllers/Api/EvaluasiController.php`: target diambil dari **versi CP yang disahkan terakhir** (bukan statistik yang terus diperbarui), klaim dibandingkan mulai `tanggal_masuk ≥ aktif_sejak`, persentil dihitung dari data klaim, dan ambang (`persen_lewat_p75`, `kenaikan_severity_iii`) dibaca dari konfigurasi.
