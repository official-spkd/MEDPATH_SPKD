# Database & Persistensi (PostgreSQL)

> Model "hybrid" yang dipakai MedPath: apa yang disimpan di PostgreSQL, apa yang tetap digenerasi di Go, bagaimana seed/load/write-through bekerja, dan jebakan yang pernah terjadi.

## Model hybrid dalam satu gambar

```text
Startup → koneksi PostgreSQL → migrasi idempoten → seed hanya bila cp_state kosong.
Request → transaksi → snapshot DB → validasi & perubahan → commit.
Galat → rollback dan HTTP 503, tanpa respons sukses semu.
DEMO_IN_MEMORY=true → demo lokal eksplisit; perubahan hilang saat restart.
```

Kenapa hybrid? Aturan domain (pengesahan, syarat, evaluasi) sudah ditulis untuk bekerja di atas struktur memori; persistensi ditambahkan tanpa menulis ulang logika itu. Data statis yang tidak pernah diubah pengguna tidak perlu tabel.

## Skema (dibuat otomatis di `persist.go`)

```sql
CREATE TABLE cp_state (
  kode TEXT PRIMARY KEY, status TEXT NOT NULL, status_sejak TEXT, menunggu_peran TEXT,
  jumlah_acuan INT NOT NULL DEFAULT 0, jumlah_rencana INT NOT NULL DEFAULT 0
);
CREATE TABLE cp_acuan    (kode TEXT, id INT, dokumen TEXT, sumber TEXT, tahun INT, catatan TEXT, PRIMARY KEY (kode, id));
CREATE TABLE cp_kriteria (kode TEXT, id INT, jenis TEXT, uraian TEXT,                          PRIMARY KEY (kode, id));
CREATE TABLE cp_rencana  (kode TEXT, id INT, jenis TEXT, severity TEXT, hari INT, sifat TEXT,
                          kode_item TEXT, nama TEXT, dosis TEXT, rute TEXT, frekuensi TEXT,     PRIMARY KEY (kode, id));
CREATE TABLE cp_riwayat  (kode TEXT, id INT, dari_status TEXT, ke_status TEXT, aksi TEXT,
                          peran TEXT, nama_user TEXT, alasan TEXT, waktu TEXT,                  PRIMARY KEY (kode, id));
CREATE TABLE padanan_kptl   (icd9cm TEXT PRIMARY KEY, kptl TEXT, oleh TEXT, pada TEXT);
CREATE TABLE padanan_snomed (icd10  TEXT PRIMARY KEY, snomed TEXT, oleh TEXT, pada TEXT, ragu BOOLEAN);
```

| Tabel | Isi | Ditulis oleh |
|---|---|---|
| `cp_state` | status & penghitung per CP | `Transisi`, tambah/hapus acuan & rencana |
| `cp_acuan` / `cp_kriteria` / `cp_rencana` | isi klinis per CP | `TambahX` / `HapusX` |
| `cp_riwayat` | audit trail pengesahan (append-only) | `Transisi` |
| `padanan_kptl` / `padanan_snomed` | pemetaan kode | `SimpanKptl`/`HapusKptl`, `SimpanSnomed`/`HapusSnomed` |

Di `cp_rencana`, kolom `kode_item` menyimpan `Rencana.Kode` (kode KFA/LOINC/ICD-9) supaya tidak bentrok dengan `kode` (kode CP).

## Yang TIDAK disimpan di DB

Statistik severity, sub-CP (diagnosis), daftar 24 CP beserta nama/kelompok/episode, 10 dokumen panduan, master KPTL & SNOMED, daftar prosedur & diagnosis klaim, serta definisi akun demo. Metadata awal berasal dari kode Go. Hash token sesi disimpan di auth_sessions (24 jam). Dokumen dan profil yang diedit disimpan di app_state; episode sumber evaluasi berada di klaim_episode.

CP baru masih memerlukan pengembangan impor data. Dokumen dapat ditambahkan melalui Dokumen Panduan; pengguna non-demo melalui Pengaturan oleh Super Admin. Hash sandi pengguna non-demo disimpan di `app_users`.

## Write-through

Helper di `persist.go` (semuanya *no-op* bila `pool == nil`):

| Helper | SQL |
|---|---|
| `simpanCPState(cp)` | `UPDATE cp_state SET status, status_sejak, menunggu_peran, jumlah_acuan, jumlah_rencana` |
| `simpanRiwayat(kode, r)` | `INSERT INTO cp_riwayat` |
| `simpanAcuanDB` / `hapusAcuanDB` | `INSERT` / `DELETE` `cp_acuan` |
| `simpanKriteriaDB` / `hapusKriteriaDB` | idem `cp_kriteria` |
| `simpanRencanaDB` / `hapusRencanaDB` | idem `cp_rencana` |
| `simpanPadananKptlDB` / `simpanPadananSnomedDB` | `INSERT … ON CONFLICT DO UPDATE` (upsert) |
| `hapusPadananKptlDB` / `hapusPadananSnomedDB` | `DELETE` |

> Semua penulisan state dan jejak pengesahan berada dalam transaksi yang sama. Galat SQL membatalkan transaksi dan dibalas dengan 503.

## Load (`loadDB`)

1. Baca `cp_state` → set status, label, penghitung pada CP yang cocok.
2. Kosongkan `Acuan`, `Kriteria`, `Rencana` semua CP, lalu isi dari tabelnya (urut `kode, id`).
3. Kosongkan riwayat, isi dari `cp_riwayat`, dan set `nextRiwayat` ke `id` terbesar + 1.
4. Ganti peta padanan dengan isi tabel padanan.
5. Muat pengguna non-demo, pengaturan administrasi, dokumen, serta episode klaim.

`Severity` dan `SubCP` tidak disentuh — tetap dari `bangunDetail`.

## Jebakan yang pernah terjadi

| Gejala | Penyebab | Perbaikan |
|---|---|---|
| Seed gagal: `duplicate key … cp_riwayat_pkey` | ID riwayat demo mulai dari 1 per CP, sedangkan PK awalnya hanya `id` | PK diubah menjadi `(kode, id)` |
| Halaman Detail CP blank untuk CP AKTIF | slice `nil` di Go → JSON `null` → `null.length` di Vue | semua slice diinisialisasi `[]T{}` |
| Setelah restart, data kembali ke awal | mode DEMO_IN_MEMORY=true sengaja diaktifkan | cek log `Terhubung ke PostgreSQL.` |

## Inspeksi manual

```bash
psql -d transcpg
\dt                                              -- daftar tabel
SELECT kode, status, status_sejak FROM cp_state ORDER BY kode;
SELECT * FROM cp_riwayat WHERE kode = 'I-4-09' ORDER BY id;
SELECT count(*) FROM cp_rencana;
```

## Backup & restore

```bash
pg_dump -d transcpg -Fc -f medpath_$(date +%F).dump     # backup
pg_restore -d transcpg --clean medpath_2026-10-07.dump  # restore
```

## Arah migrasi berikutnya

Pengguna dan dokumen sudah dapat dikelola. Pengembangan berikutnya:

1. Pertahankan `app_users` (PBKDF2-HMAC-SHA256) dan `auth_sessions`; tambahkan pembatasan laju login.
2. Tabel `clinical_pathway`, `klaim_episode`, `dokumen_panduan`, master kode — rujuk skema migrasi Laravel di `_legacy/backend-laravel/database/migrations`.
3. Ganti generator demo di `cpg.go`/`detail.go` dengan query.
4. Kenalkan alat migrasi berversi (mis. `goose` atau `golang-migrate`) menggantikan `CREATE TABLE IF NOT EXISTS`.
