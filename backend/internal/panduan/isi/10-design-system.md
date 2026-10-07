# Design System & UI

> Token warna, tipografi, kelas utilitas, cara kerja mode gelap, dan aturan visual MedPath — supaya halaman baru terlihat menyatu dengan yang lama.

## Prinsip

- **Tata letak mengikuti MedClaim** (sidebar gradasi, topbar ber-blur, hero gelap, kartu KPI, tabel bersih) — **warnanya tidak**. MedPath memakai **salmon/merah**; gradasi selalu bergerak ke merah yang lebih muda atau lebih tua, tidak pernah ke emas.
- Semua warna diambil dari **CSS variable** di `frontend/src/assets/tokens.css`. Jangan menulis hex langsung di komponen kecuali di dalam SVG yang memang butuh (dan beri komentar).
- Warna status (hijau/merah/kuning/ungu/biru) bersifat **fungsional** — dipakai untuk makna data, bukan dekorasi.

## Token warna utama

| Token | Terang | Kegunaan |
|---|---|---|
| `--brand` | `#f2695e` | warna merek salmon-merah, tombol utama |
| `--brand-hover` | `#e8544a` | hover tombol |
| `--brand-aktif` | `#d13b30` | teks aksen, kode CP, ikon aktif |
| `--brand-light` | `#fdeceb` | latar ikon/badge lembut |
| `--brand-gelap` | `#9e2b22` | teks di atas latar terang merek |
| `--gradient-brand` | `#3a0f0b → #8a241a → #e85749` | kotak logo, avatar |
| `--gradient-hero` | radial salmon + `#290d0a → #6b231b` | hero Dashboard, panel Login |
| `--sidebar-grad` | `#ffc1b6 → #fb988b → #f2695e` | latar sidebar (tetap di mode gelap) |
| `--teal` | `#0f766e` | aksen sekunder kecil (garis CP Aktif di grafik) |
| `--bg` / `--surface` | `#fbf7f5` / `#ffffff` | latar halaman / kartu |
| `--border` / `--border-kuat` | `#f0e4e0` / `#e4cfc9` | garis |
| `--text` / `--text-2` / `--text-3` | `#2a2422` / `#5c514e` / `#a89a95` | teks utama / sekunder / redup |

Warna status: `--hijau` `#059669`, `--merah` `#dc2626`, `--kuning` `#c2710d`, `--ungu` `#7c3aed`, `--info` `#0284c7`, masing-masing punya pasangan `-bg` untuk latar lembut.

Pemetaan status CP → warna (`StatusPill.vue`): AKTIF hijau · MENUNGGU_DIREKTUR kuning · REVIEW_KOMITE ungu · REVIEW_TIM_CP biru · REVISI merah · DRAFT netral.

## Mode gelap

`tokens.css` mendefinisikan ulang token latar, teks, garis, dan `-bg` di bawah selektor **`html.dark`**. `stores/tema.ts` cukup men-toggle class itu, dan skrip di `index.html` menerapkannya sebelum render. Karena semua komponen membaca variable, **halaman baru otomatis mendukung mode gelap** selama tidak memakai hex langsung.

## Tipografi

| Token | Font | Dipakai untuk |
|---|---|---|
| `--font-judul` | Poppins 500–800 | `h1`–`h5`, angka besar KPI |
| `--font-teks` | Open Sans 400–700 | teks isi |

Tambahkan kelas `.angka` pada angka agar memakai *tabular numerals* (lebar digit seragam di tabel), dan `.mono` untuk kode (ICD, KFA, kode CP).

## Kelas utilitas (`base.css`)

| Kelas | Hasil |
|---|---|
| `.card` | kartu putih, garis, sudut 14px, bayangan lembut |
| `.btn` + `.btn-brand` / `.btn-outline` / `.btn-ghost` | tombol 40px; varian utama/garis/transparan |
| `.badge` | pil kecil; beri `background`/`color` sesuai makna |
| `.muted`, `.dimmed` | teks sekunder / redup |
| `.mono`, `.angka` | font kode / angka tabular |

Radius: `--radius` 14px (kartu), `--radius-sm` 10px (input, tombol), `--radius-pill` 999px. Bayangan: `--shadow`, `--shadow-md` (hover), `--shadow-lg`.

## Pola halaman standar

```vue
<template>
  <div>
    <PageHeader judul="Judul Halaman" demo sub="Satu kalimat penjelasan tujuan halaman.">
      <template #aksi><button class="btn btn-brand">Aksi Utama</button></template>
    </PageHeader>

    <div class="card panel"> … </div>
  </div>
</template>
```

- Kartu ringkasan angka → grid `repeat(auto-fit, minmax(150px, 1fr))`.
- Tabel → header kecil huruf kapital `--text-3`, baris hover `--bg`, kode berwarna `--brand-aktif` + `.mono`.
- Status kosong → kotak tengah dengan ikon dalam lingkaran `--brand-light`.
- Galat → kotak `--merah-bg` teks `--merah`.

## Responsif

Breakpoint yang dipakai: **1180px** (hero dashboard menumpuk, KPI 3 kolom), **920px** (panel Login disembunyikan, kolom detail CP menumpuk, pencarian topbar disembunyikan), **860px/720px** (grid 2 → 1 kolom).

## Ikon

Pakai `lucide-vue-next`, impor per ikon (`import { Stamp } from 'lucide-vue-next'`) supaya *tree-shaking* bekerja. Ukuran standar 16–20 px, `stroke-width` 2–2.2.
