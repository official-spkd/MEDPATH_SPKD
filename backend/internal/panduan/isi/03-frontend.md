# Frontend (Vue 3 + Vite)

> Cara SPA dirakit: bootstrap, router & guard, store Pinia, klien API, layout, komponen, dan halaman — beserta alasan di balik tiap pilihan.

## Bootstrap

**`index.html`** memuat font Poppins + Open Sans dari Google Fonts dan menjalankan skrip kecil **sebelum** Vue dimuat:

```html
<script>
  var t = localStorage.getItem('transcpg-tema')
  if (t === 'dark' || (!t && matchMedia('(prefers-color-scheme: dark)').matches))
    document.documentElement.classList.add('dark')
</script>
```

Tujuannya mencegah "kedip putih" saat pengguna memakai mode gelap.

**`src/main.ts`** membuat aplikasi, memasang Pinia dan router, mengimpor `base.css`, dan memasang **error handler global**:

```ts
app.config.errorHandler = (err, _inst, info) => {
  console.error('[VUE ERR]', info, err?.message, err?.stack)
}
```

Kalau sebuah halaman tiba-tiba kosong, buka DevTools → Console dan cari `[VUE ERR]` — biasanya langsung menunjuk baris template yang gagal.

## Router & guard (`src/router/index.ts`)

- Memakai **`createWebHashHistory()`** → URL berbentuk `/#/library`. Keuntungannya: build statis bisa disajikan server apa pun tanpa aturan *rewrite*.
- Semua halaman dimuat **lazy** (`() => import('@/pages/…')`) sehingga tiap halaman jadi chunk terpisah.
- Halaman di dalam aplikasi adalah anak dari rute `/` yang memakai `AppLayout.vue`.

Field `meta` yang dikenali guard:

| meta | Arti |
|---|---|
| `publik: true` | boleh dibuka tanpa login (hanya `/login`) |
| `judul` | teks breadcrumb di Topbar |
| `admin: true` | hanya ADMIN_RS, SYSTEM_ADMIN, SUPER_ADMIN |
| `superAdmin: true` | hanya SUPER_ADMIN |

Urutan kerja `router.beforeEach`:

1. Bila sesi belum dicek (`auth.siap === false`), panggil `auth.muatProfil()` → `GET /auth/me`.
2. Belum login & rute tidak publik → lempar ke `/login?ke=<tujuan>`; setelah login, pengguna dikembalikan ke `ke`.
3. Sudah login tapi membuka `/login` → ke `/`.
4. `meta.admin` / `meta.superAdmin` tidak terpenuhi → dialihkan.

> Guard di frontend hanya untuk kenyamanan. **Batas keamanan sesungguhnya ada di backend** — misalnya `/panduan` tetap membalas `403` untuk selain SUPER_ADMIN walau seseorang mengakali router.

## Store Pinia

### `stores/auth.ts`

| Bagian | Isi |
|---|---|
| state | `pengguna` (id, nama, email, peran, peran_label), `token`, `siap` |
| getter | `masuk` (ada token), `isAdmin`, `isSuperAdmin` |
| action | `login(email, password)`, `muatProfil()`, `logout()`, `bersihkan()` |

Token disimpan di `localStorage` dengan kunci **`transcpg-token`**. Jika `/auth/me` gagal (mis. backend baru restart sehingga token tak dikenal), `bersihkan()` menghapus token dan pengguna diarahkan ke login.

### `stores/tema.ts`

Menyimpan `mode` (`light`/`dark`) di `localStorage` kunci **`transcpg-tema`**, dan `terapkan()` men-toggle class `dark` pada `<html>`. Semua warna dibaca dari CSS variable, jadi mengganti class sudah cukup untuk mengganti tema.

## Klien API (`src/lib/api.ts`)

Satu fungsi generik untuk semua panggilan:

```ts
const data = await api<Detail>(`cp/${kode}`)
await api(`cp/${kode}/transisi`, { method: 'POST', body: JSON.stringify({ aksi: 'SAHKAN' }) })
```

Yang dilakukannya:

1. Prefiks `BASE = '/api/v1'`.
2. Header `Accept: application/json`; `Content-Type: application/json` bila ada body.
3. `Authorization: Bearer <token>` bila sudah login.
4. **Cache-buster** untuk GET: menambahkan `?_=<timestamp>` + `cache: 'no-store'`. Ini sengaja — sebelumnya browser sempat menyajikan detail CP lama setelah status berubah.
5. Status `204` → `undefined`. Status non-2xx → melempar **`GalatApi(status, message)`** dengan pesan dari field `message` backend.

Pola penanganan galat di halaman:

```ts
try {
  d.value = await api<Detail>(`cp/${kode.value}`)
} catch (e) {
  galat.value = e instanceof GalatApi ? e.message : 'Gagal memuat detail CP.'
}
```

## Utilitas format (`src/lib/format.ts`)

| Fungsi / konstanta | Contoh |
|---|---|
| `angka(2287)` | `2.287` (format id-ID) |
| `rupiahRingkas(551800000)` | `Rp 551,8 jt` |
| `persen(3, 24)` | `12,5%` |
| `inisial('Super Admin (demo)')` | `SA` |
| `sapaan()` | Selamat pagi/siang/sore/malam sesuai jam |
| `LABEL_STATUS`, `LABEL_KATEGORI_BUTIR`, `LABEL_SUMBER` | peta kode → label Indonesia |

## Layout (`layouts/AppLayout.vue`)

Menyusun **Sidebar** (kiri, sticky) + **Topbar** (atas, sticky, blur) + `<router-view>`. Layout juga memuat jumlah antrean Approval untuk badge sidebar, dan menyediakan fungsi lewat `provide('refreshAntrean', …)`. Halaman yang mengubah status CP memanggilnya lewat `inject('refreshAntrean')` agar badge langsung diperbarui.

## Komponen

| Komponen | Fungsi | Props penting |
|---|---|---|
| `Sidebar.vue` | menu grup *Utama* & *Referensi*, badge antrean, kartu akun, tombol Keluar | `antrean` |
| `Topbar.vue` | breadcrumb dari `meta.judul`, pencarian, filter RS/Layanan, toggle tema, notifikasi | — |
| `LogoSpkd.vue` | logo MedPath | `terang`, `ringkas` |
| `PageHeader.vue` | judul halaman, subjudul, badge "Data Demo", slot `aksi` | `judul`, `sub`, `demo` |
| `StatCard.vue` | kartu KPI berikon + garis aksen; jadi tautan bila diberi `to` | `label`, `nilai`, `icon`, `warna`, `keterangan`, `to` |
| `StatusPill.vue` | pil warna untuk status CP | `status` |
| `Gauge.vue` | busur 270° skor readiness | `skor`, `maks` |
| `DonutChart.vue` | donat + legenda persentase | `segmen[]`, `pusatAngka`, `pusatLabel` |
| `AreaChart.vue` | area episode + garis putus-putus CP aktif, dua skala | `data[]` |

Catatan jujur: kolom pencarian dan dropdown *Semua RS / Semua Layanan* di Topbar saat ini **dekoratif** (belum terhubung ke data).

### Teknik grafik SVG

Gauge dan donat memakai `pathLength="100"` pada `<circle>` sehingga `stroke-dasharray` bisa ditulis dalam persen:

```html
<circle r="58" pathLength="100" stroke-dasharray="75 25" transform="rotate(135 70 70)" />
```

Busur 270° = 75 dari 100; rotasi 135° menaruh celahnya di bawah. Progres = `skor/maks × 75`.

## Halaman

| Halaman | Hal yang perlu diketahui |
|---|---|
| `Login.vue` | panel merek + form; chip akun demo memanggil `login()` langsung |
| `Dashboard.vue` | satu panggilan `GET /dashboard`, lalu dipecah ke hero readiness, 6 KPI, tren, tindakan prioritas, donat status, bar kelompok, tabel prioritas |
| `CpLibrary.vue` | penyaringan dilakukan di klien (pencarian + status) |
| `CpDetail.vue` | memuat detail & riwayat lewat `watch(kode, muat, { immediate: true })`; tab Ringkasan/Acuan/Rencana/Kriteria/Riwayat; panel Pengesahan + konfirmasi; form edit isi tampil hanya bila `boleh_ubah` |
| `Approval.vue` | daftar CP yang sedang diproses atau perlu revisi |
| `CpAktif.vue` | kartu CP berstatus AKTIF |
| `Evaluasi.vue` | daftar lipat; detail evaluasi dimuat saat baris dibuka lalu di-cache |
| `Padanan.vue` | **satu file untuk dua rute**; `route.name === 'snomed'` menentukan endpoint & label |
| `Pengaturan.vue` | kartu administrasi; kartu Panduan hanya untuk Super Admin |
| `PanduanSuperAdmin.vue` | daftar isi, pencarian, render Markdown, unduh `.md` |

### Kenapa `watch(..., { immediate: true })` dan bukan `onMounted`

`CpDetail` dipakai ulang oleh vue-router saat berpindah dari satu kode CP ke kode lain (komponen yang sama, parameter berbeda), sehingga `onMounted` tidak terpanggil lagi. `watch` pada `kode` memuat ulang data setiap parameter berubah, sekaligus saat pertama dibuka.

### Pola edit isi di `CpDetail`

Semua aksi tambah/hapus memakai satu pembungkus:

```ts
async function kirimEdit(fn: () => Promise<Detail>, reset?: () => void) {
  pesanEdit.value = ''
  try { d.value = await fn(); reset?.() }
  catch (e) { pesanEdit.value = e instanceof GalatApi ? e.message : 'Gagal menyimpan.' }
}
```

Backend selalu membalas **detail CP terbaru** setelah mutasi, sehingga daftar syarat ikut ter-update tanpa permintaan kedua.

## Konfigurasi build (`vite.config.ts`)

- Alias `@` → `src`.
- `server.port = 5173`.
- `server.proxy['/api']` → `BACKEND_URL` atau `http://127.0.0.1:8080`.

Hasil `npm run build` berada di `frontend/dist/` dan berisi HTML + aset bernama hash; aman di-cache lama oleh CDN.
