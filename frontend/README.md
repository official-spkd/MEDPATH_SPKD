# MedPath — Frontend (Vue + Vite)

SPA Vue 3 + Vite untuk Clinical Pathway Generator SPKD. UI/UX mengikuti gaya **MedClaim**
(salmon/merah, Poppins/Open Sans, sidebar gradient, dark mode).

## Menjalankan

```bash
cd frontend
npm install
npm run dev        # http://localhost:5173
```

Backend Go harus jalan di `:8080` (lihat `../backend`). Dev server mem-proxy `/api` → backend.
Arahkan ke backend lain:

```bash
BACKEND_URL=http://127.0.0.1:9090 npm run dev
```

Build produksi:

```bash
npm run build      # keluar ke dist/
npm run preview
```

## Stack

- **Vue 3** (`<script setup>`, TypeScript) + **Vite 6**
- **Vue Router** (hash routing, `#/`)
- **Pinia** (store `auth`, `tema`)
- **lucide-vue-next** (ikon)
- Chart (gauge, donut, area) digambar sendiri sebagai **SVG** — tanpa library chart.
- Styling: CSS variabel (design token) di `src/assets/tokens.css`, tanpa framework CSS.

## Struktur

```
src/
├── assets/          tokens.css (design system) + base.css
├── components/      LogoSpkd, Sidebar, Topbar, StatCard, StatusPill,
│                    Gauge, DonutChart, AreaChart, PageHeader
├── layouts/         AppLayout (sidebar + topbar + <router-view>)
├── pages/           Login, Dashboard, CpLibrary, Approval, CpAktif,
│                    Evaluasi, Dokumen, Padanan, Pengaturan
├── stores/          auth.ts, tema.ts
├── lib/             api.ts (fetch + token), format.ts (angka/label)
└── router/          index.ts (guard auth + admin)
```

## Design token

Semua warna/jari-jari/font ada di `src/assets/tokens.css` — light di `:root`, dark di `html.dark`.
Brand amber `#f0a640`, teal `#0f766e`, cream `#faf8f4` / dark `#14100a`.
