# Alur Pengesahan & Aturan Domain

> Siklus hidup sebuah Clinical Pathway, siapa boleh melakukan apa di status apa, kapan isi CP terkunci, dan tujuh syarat kelengkapan sebelum CP boleh AKTIF. Semua aturan ini ada di `backend/internal/store/pengesahan.go`.

## Status CP

| Kode | Label | Arti |
|---|---|---|
| `DRAFT` | Draft | sedang disusun |
| `REVIEW_TIM_CP` | Review Tim CP | diajukan, ditinjau Tim CP |
| `REVIEW_KOMITE` | Review Komite Medik | lolos Tim CP, ditinjau Komite |
| `MENUNGGU_DIREKTUR` | Menunggu Direktur | lolos Komite, menunggu pengesahan |
| `AKTIF` | Aktif | disahkan; dipakai memandu pelayanan |
| `REVISI` | Perlu Revisi | dikembalikan dari salah satu tahap |

"Sedang diproses" (`sedangDiproses`) = `REVIEW_TIM_CP`, `REVIEW_KOMITE`, `MENUNGGU_DIREKTUR`. Antrean **Approval** = sedang diproses **+** `REVISI`.

## Diagram alur

```text
Jalur maju:
  DRAFT ─AJUKAN─▶ REVIEW_TIM_CP ─SETUJUI─▶ REVIEW_KOMITE ─SETUJUI─▶ MENUNGGU_DIREKTUR ─SAHKAN─▶ AKTIF

Jalur balik (dari tahap mana pun setelah diajukan, termasuk AKTIF):
  REVIEW_TIM_CP │ REVIEW_KOMITE │ MENUNGGU_DIREKTUR │ AKTIF ─KEMBALIKAN─▶ REVISI

Pengajuan ulang:
  REVISI ─AJUKAN─▶ REVIEW_TIM_CP          (mulai lagi dari Tim CP)
```

## Kelompok peran

Didefinisikan sebagai fungsi agar mudah dipakai ulang:

| Nama di kode | Anggota |
|---|---|
| `peranAdmin` | ADMIN_RS, SYSTEM_ADMIN, SUPER_ADMIN |
| `penyusun()` | TIM_CP, KOORDINATOR_CP, DPJP, DOKTER + `peranAdmin` |
| `peranTimCp()` | TIM_CP, KOORDINATOR_CP + `peranAdmin` |

## Matriks transisi (`transisiDari`)

| Dari status | Aksi | Ke status | Peran yang boleh |
|---|---|---|---|
| DRAFT, REVISI | AJUKAN | REVIEW_TIM_CP | `penyusun()` |
| REVIEW_TIM_CP | SETUJUI | REVIEW_KOMITE | `peranTimCp()` |
| REVIEW_TIM_CP | KEMBALIKAN | REVISI | `peranTimCp()` |
| REVIEW_KOMITE | SETUJUI | MENUNGGU_DIREKTUR | KOMITE_MEDIK + `peranAdmin` |
| REVIEW_KOMITE | KEMBALIKAN | REVISI | KOMITE_MEDIK + `peranAdmin` |
| MENUNGGU_DIREKTUR | SAHKAN | AKTIF | DIREKTUR, SYSTEM_ADMIN, SUPER_ADMIN |
| MENUNGGU_DIREKTUR | KEMBALIKAN | REVISI | DIREKTUR, SYSTEM_ADMIN, SUPER_ADMIN |
| AKTIF | KEMBALIKAN | REVISI | KOMITE_MEDIK, DIREKTUR + `peranAdmin` |

Perhatikan: **ADMIN_RS tidak bisa SAHKAN** — pengesahan tetap hak Direktur (atau admin sistem).

Fungsi terkait:

- `aksiUntuk(peran, status)` → aksi yang boleh dilakukan peran itu sekarang (dipakai untuk menampilkan tombol di UI).
- `peranDitunggu(status)` → label peran yang ditunggu (aksi selain KEMBALIKAN, tanpa admin) — tampil sebagai "Menunggu tindakan: Direktur".

## Cara `Transisi()` memutuskan

```text
1. Kunci store (Lock).
2. CP ada?                                   tidak → "CP tidak ditemukan."
3. Aksi tersedia dari status sekarang?       tidak → "Aksi X tidak tersedia pada status Y."
4. Peran user termasuk peran aksi itu?       tidak → "Peran Anda tidak berwenang…"
5. Tujuan = AKTIF? → semua syarat menghambat terpenuhi?
                                             tidak → "Belum bisa disahkan: <syarat> belum terpenuhi."
6. Ubah status, status_sejak = hari ini, menunggu_peran baru.
7. Tambah RiwayatItem (dari, ke, aksi, peran, nama user, alasan, waktu).
8. Write-through: cp_state + cp_riwayat.
```

Penolakan dikembalikan sebagai HTTP `422` dengan pesan di atas.

> **Celah yang perlu diketahui:** UI mewajibkan *alasan* saat KEMBALIKAN, tetapi backend belum memvalidasinya. Bila suatu saat ada klien lain (mis. integrasi), tambahkan pengecekan `alasan != ""` untuk aksi KEMBALIKAN di `Transisi`.

## Kapan isi CP boleh diubah (`pengubahIsi`)

Isi = acuan, kriteria, rencana.

| Status | Acuan & Kriteria | Rencana |
|---|---|---|
| DRAFT, REVISI | `penyusun()` | `penyusun()` + APOTEKER |
| REVIEW_TIM_CP | `peranTimCp()` | `peranTimCp()` |
| REVIEW_KOMITE, MENUNGGU_DIREKTUR, AKTIF | **terkunci** | **terkunci** |

Alasan penguncian: yang disahkan Direktur harus **sama persis** dengan yang ditinjau Komite, dan CP aktif sedang dipakai melayani pasien — perubahan harus lewat KEMBALIKAN → REVISI → diajukan ulang.

`Detail()` mengirim `boleh_ubah: {acuan, rencana, kriteria}` untuk user yang meminta; frontend hanya menampilkan tombol tambah/hapus bila bernilai `true`. Backend tetap memeriksa ulang di `TambahAcuan`/`HapusRencana`/dst. dan membalas `422` "Isi CP terkunci…" bila dilanggar.

## Tujuh syarat kelengkapan (`periksaSyarat`)

| # | Kode | Judul | Menghambat SAHKAN? | Penanggung jawab | Aturan di kode |
|---|---|---|---|---|---|
| 1 | `ACUAN_CP` | Acuan standar CP — minimal satu dokumen | **Ya** | Tim CP · KSM | `len(acuan) > 0` |
| 2 | `RENCANA_SEVERITY` | Rencana isi klinis tiap severity dengan data cukup | **Ya** | Tim CP · DPJP · Farmasi | tiap severity dengan **≥ 10 episode** wajib punya ≥ 1 rencana |
| 3 | `ACUAN_SUB_CP` | Acuan tiap sub-CP | Tidak | Tim CP · KSM terkait | setiap sub-CP `ada_acuan` |
| 4 | `PADANAN_KPTL` | Padanan KPTL tiap prosedur | Tidak | Tim Koding | *demo:* terpenuhi bila CP punya rencana |
| 5 | `TARIF_SEVERITY` | Tarif INA-CBG tiap severity | Tidak | Admin · data tarif | *demo:* terpenuhi bila CP punya rencana |
| 6 | `PADANAN_SNOMED` | Padanan SNOMED-CT tiap diagnosis | Tidak | Tim Koding | *demo:* terpenuhi bila CP punya acuan |
| 7 | `ICD10_MASTER` | Kode ICD-10 sesuai master | Tidak | Tim Koding | *demo:* selalu terpenuhi |

Semua syarat **selalu dihitung** (bukan berhenti di yang pertama gagal) supaya UI bisa menampilkan seluruh kekurangan sekaligus. Hanya syarat 1 dan 2 yang memblokir SAHKAN.

Syarat 4–7 masih heuristik demo. Untuk versi produksi, rujuk logika asli di `_legacy/backend-laravel/app/Cpg/PemeriksaSyarat.php`: KPTL dicek dari prosedur klaim yang belum punya padanan, tarif dari tabel tarif per kelas, SNOMED dari padanan berstatus bukan RAGU, ICD-10 dari master.

## Riwayat (audit trail)

Setiap transisi menambah satu `RiwayatItem` dan **tidak pernah diubah atau dihapus**. Riwayat awal untuk data demo dibangkitkan oleh `bangunRiwayatAwal`, yang menelusuri jalur DRAFT → … sampai status CP saat ini (dan menambahkan langkah KEMBALIKAN untuk CP berstatus REVISI).

Di database, kunci primer riwayat adalah **`(kode, id)`** — bukan `id` saja — karena riwayat demo diberi nomor mulai 1 per CP.

## Mengubah aturan dengan aman

1. Ubah `transisiDari` / `pengubahIsi` / `periksaSyarat` di `pengesahan.go`.
2. Jangan lupa label: `labelAksi` (teks tombol) dan `labelPeran`.
3. Bila menambah **status baru**: tambahkan juga di `statusLabel` (`cpg.go`), `sedangDiproses` bila relevan, `bangunRiwayatAwal`, serta di frontend `LABEL_STATUS` (`format.ts`) dan peta warna `StatusPill.vue`, plus urutan & warna donat di `Dashboard.vue`.
4. Jalankan `go test ./...` dan uji manual dengan beberapa akun demo.
