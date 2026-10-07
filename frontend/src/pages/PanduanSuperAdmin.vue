<script setup lang="ts">
import { ref, computed, onMounted, watch, nextTick } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { Marked } from 'marked'
import { ArrowLeft, ArrowRight, Search, Download, ShieldCheck, CircleAlert, BookMarked } from 'lucide-vue-next'
import { api, GalatApi } from '@/lib/api'

interface Bab {
  id: string
  urutan: number
  judul: string
  ringkasan: string
  isi: string
}

const route = useRoute()
const router = useRouter()

const daftar = ref<Bab[]>([])
const galat = ref('')
const memuat = ref(true)
const cari = ref('')
const isiEl = ref<HTMLElement | null>(null)

onMounted(async () => {
  try {
    const r = await api<{ data: Bab[] }>('panduan')
    daftar.value = r.data
  } catch (e) {
    galat.value = e instanceof GalatApi ? e.message : 'Gagal memuat panduan.'
  } finally {
    memuat.value = false
  }
})

// --- Markdown: heading ber-id & blok kode dengan tombol salin ---
function escapeHtml(s: string) {
  return s.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;').replace(/"/g, '&quot;')
}
function slug(s: string) {
  return s
    .toLowerCase()
    .replace(/<[^>]+>/g, '')
    .replace(/[^a-z0-9\s-]/g, '')
    .trim()
    .replace(/\s+/g, '-')
}
const md = new Marked({ gfm: true })
md.use({
  renderer: {
    heading({ tokens, depth, text }) {
      return `<h${depth} id="h-${slug(text)}">${this.parser.parseInline(tokens)}</h${depth}>\n`
    },
    code({ text, lang }) {
      const label = lang ? `<span class="kode-lang">${escapeHtml(lang)}</span>` : ''
      return `<div class="kode-blok">${label}<button type="button" class="salin" data-salin>Salin</button><pre><code>${escapeHtml(text)}</code></pre></div>\n`
    },
  },
})

// --- Pemilihan bab (disimpan di ?bab= agar bisa dibagikan) ---
const babAktif = computed<Bab | undefined>(() => {
  const id = route.query.bab as string | undefined
  return daftar.value.find((b) => b.id === id) ?? daftar.value[0]
})
const indeks = computed(() => daftar.value.findIndex((b) => b.id === babAktif.value?.id))
const sebelumnya = computed(() => (indeks.value > 0 ? daftar.value[indeks.value - 1] : undefined))
const berikutnya = computed(() =>
  indeks.value >= 0 && indeks.value < daftar.value.length - 1 ? daftar.value[indeks.value + 1] : undefined,
)

function pilih(b: Bab) {
  router.replace({ query: { bab: b.id } })
}
watch(
  () => babAktif.value?.id,
  async () => {
    await nextTick()
    window.scrollTo({ top: 0 })
  },
)

const html = computed(() => (babAktif.value ? (md.parse(babAktif.value.isi) as string) : ''))

// Sub-judul (##) bab aktif untuk navigasi cepat
const subJudul = computed(() =>
  (babAktif.value?.isi.match(/^## .+$/gm) ?? []).map((l) => {
    const t = l.replace(/^## /, '')
    return { teks: t.replace(/[`*]/g, ''), id: `h-${slug(t)}` }
  }),
)
function lompat(id: string) {
  document.getElementById(id)?.scrollIntoView({ behavior: 'smooth', block: 'start' })
}

// --- Pencarian ---
const hasil = computed(() => {
  const q = cari.value.trim().toLowerCase()
  if (!q) return daftar.value
  return daftar.value.filter((b) => (b.judul + ' ' + b.ringkasan + ' ' + b.isi).toLowerCase().includes(q))
})
function jumlahCocok(b: Bab) {
  const q = cari.value.trim().toLowerCase()
  if (!q) return 0
  return (b.isi.toLowerCase().match(new RegExp(q.replace(/[.*+?^${}()|[\]\\]/g, '\\$&'), 'g')) ?? []).length
}

// Clipboard API butuh konteks aman + fokus dokumen; fallback ke textarea + execCommand.
async function salinTeks(teks: string): Promise<boolean> {
  try {
    await navigator.clipboard.writeText(teks)
    return true
  } catch {
    const ta = document.createElement('textarea')
    ta.value = teks
    ta.setAttribute('readonly', '')
    ta.style.position = 'fixed'
    ta.style.opacity = '0'
    document.body.appendChild(ta)
    ta.select()
    const ok = document.execCommand('copy')
    ta.remove()
    return ok
  }
}

// --- Salin blok kode (event delegation pada konten v-html) ---
async function klikIsi(e: MouseEvent) {
  const btn = (e.target as HTMLElement).closest('[data-salin]') as HTMLButtonElement | null
  if (!btn) return
  const kode = btn.parentElement?.querySelector('code')?.textContent ?? ''
  btn.textContent = (await salinTeks(kode)) ? 'Tersalin ✓' : 'Gagal'
  setTimeout(() => (btn.textContent = 'Salin'), 1500)
}

// --- Unduh Markdown ---
function keMarkdown(b: Bab) {
  return `# ${b.urutan}. ${b.judul}\n\n> ${b.ringkasan}\n\n${b.isi}\n`
}
function unduh(nama: string, teks: string) {
  const url = URL.createObjectURL(new Blob([teks], { type: 'text/markdown;charset=utf-8' }))
  const a = document.createElement('a')
  a.href = url
  a.download = nama
  a.click()
  URL.revokeObjectURL(url)
}
const unduhBab = () => babAktif.value && unduh(`MedPath_Panduan_${String(babAktif.value.urutan).padStart(2, '0')}_${babAktif.value.id}.md`, keMarkdown(babAktif.value))
const unduhSemua = () =>
  unduh('MedPath_Panduan_Lengkap_Super_Admin.md', '# Panduan Lengkap Super Admin — MedPath\n\n' + daftar.value.map(keMarkdown).join('\n---\n\n'))
</script>

<template>
  <div>
    <button class="btn btn-ghost btn-sm kembali" @click="router.push('/pengaturan')"><ArrowLeft :size="16" /> Pengaturan</button>

    <div class="kepala">
      <div>
        <div class="judul-baris">
          <h1>Panduan Lengkap Super Admin</h1>
          <span class="badge label-sa"><ShieldCheck :size="13" /> Khusus Super Admin</span>
        </div>
        <p class="sub">Dokumentasi teknis MedPath — arsitektur, frontend, backend, database, API, keamanan, deploy, dan troubleshooting.</p>
      </div>
      <button v-if="daftar.length" class="btn btn-outline" @click="unduhSemua"><Download :size="16" /> Unduh semua (.md)</button>
    </div>

    <div v-if="galat" class="galat-box"><CircleAlert :size="18" /> {{ galat }}</div>
    <div v-else-if="memuat" class="memuat">Memuat panduan…</div>

    <div v-else-if="babAktif" class="tata">
      <!-- Daftar isi -->
      <aside class="toc card">
        <div class="cari">
          <Search :size="16" class="cari-ikon" />
          <input v-model="cari" placeholder="Cari di panduan…" />
        </div>
        <p v-if="cari" class="cari-info">{{ hasil.length }} bab cocok</p>
        <nav class="toc-list">
          <button
            v-for="b in hasil"
            :key="b.id"
            class="toc-item"
            :class="{ aktif: b.id === babAktif.id }"
            @click="pilih(b)"
          >
            <span class="no angka">{{ String(b.urutan).padStart(2, '0') }}</span>
            <span class="toc-judul">{{ b.judul }}</span>
            <span v-if="cari && jumlahCocok(b)" class="cocok angka">{{ jumlahCocok(b) }}</span>
          </button>
          <p v-if="!hasil.length" class="kosong-kecil">Tidak ada bab yang memuat "{{ cari }}".</p>
        </nav>
      </aside>

      <!-- Isi bab -->
      <article class="konten card">
        <header class="bab-kepala">
          <p class="bab-no">Bab {{ babAktif.urutan }} dari {{ daftar.length }}</p>
          <h2 class="bab-judul">{{ babAktif.judul }}</h2>
          <p class="bab-ringkas">{{ babAktif.ringkasan }}</p>
          <div class="bab-aksi">
            <button class="btn btn-ghost btn-sm" @click="unduhBab"><Download :size="14" /> Unduh bab ini</button>
          </div>
          <div v-if="subJudul.length" class="sub-nav">
            <button v-for="s in subJudul" :key="s.id" class="sub-chip" @click="lompat(s.id)">{{ s.teks }}</button>
          </div>
        </header>

        <div ref="isiEl" class="prosa" v-html="html" @click="klikIsi" />

        <footer class="navbab">
          <button v-if="sebelumnya" class="nav-btn" @click="pilih(sebelumnya)">
            <ArrowLeft :size="16" />
            <span><small>Sebelumnya</small>{{ sebelumnya.judul }}</span>
          </button>
          <span v-else />
          <button v-if="berikutnya" class="nav-btn kanan" @click="pilih(berikutnya)">
            <span><small>Berikutnya</small>{{ berikutnya.judul }}</span>
            <ArrowRight :size="16" />
          </button>
        </footer>
      </article>
    </div>

    <div v-else class="card kosong"><BookMarked :size="22" /> Panduan belum berisi bab.</div>
  </div>
</template>

<style scoped>
.kembali {
  margin-bottom: 10px;
}
.btn-sm {
  height: 32px;
  padding: 0 10px;
  font-size: 13px;
}
.kepala {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 16px;
  flex-wrap: wrap;
  margin-bottom: 20px;
}
.judul-baris {
  display: flex;
  align-items: center;
  gap: 12px;
  flex-wrap: wrap;
}
.kepala h1 {
  font-size: 26px;
  font-weight: 800;
  letter-spacing: -0.5px;
}
.sub {
  margin-top: 6px;
  color: var(--text-2);
  font-size: 14px;
  max-width: 72ch;
}
.label-sa {
  background: var(--brand-aktif);
  color: #fff;
}
.galat-box {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 14px 16px;
  border-radius: var(--radius);
  background: var(--merah-bg);
  color: var(--merah);
  font-weight: 600;
}
.memuat,
.kosong {
  padding: 60px;
  text-align: center;
  color: var(--text-3);
  display: flex;
  gap: 10px;
  justify-content: center;
  align-items: center;
}

.tata {
  display: grid;
  grid-template-columns: 290px 1fr;
  gap: 18px;
  align-items: start;
}

/* Daftar isi */
.toc {
  position: sticky;
  top: 84px;
  padding: 14px;
  max-height: calc(100dvh - 104px);
  display: flex;
  flex-direction: column;
}
.cari {
  display: flex;
  align-items: center;
  gap: 8px;
  height: 38px;
  padding: 0 10px;
  background: var(--bg);
  border: 1px solid var(--border);
  border-radius: var(--radius-sm);
}
.cari-ikon {
  color: var(--text-3);
  flex-shrink: 0;
}
.cari input {
  flex: 1;
  min-width: 0;
  border: none;
  outline: none;
  background: transparent;
  font-size: 13.5px;
}
.cari-info {
  font-size: 12px;
  color: var(--text-3);
  margin: 8px 2px 0;
}
.toc-list {
  margin-top: 10px;
  overflow-y: auto;
  display: flex;
  flex-direction: column;
  gap: 2px;
}
.toc-item {
  display: flex;
  align-items: center;
  gap: 10px;
  width: 100%;
  text-align: left;
  padding: 8px 10px;
  border: none;
  border-radius: 9px;
  background: transparent;
  color: var(--text-2);
  font-size: 13.5px;
  font-weight: 600;
}
.toc-item:hover {
  background: var(--bg);
  color: var(--text);
}
.toc-item.aktif {
  background: var(--brand-light);
  color: var(--brand-aktif);
}
.no {
  font-size: 11px;
  font-weight: 800;
  color: var(--text-3);
  width: 20px;
  flex-shrink: 0;
}
.toc-item.aktif .no {
  color: var(--brand-aktif);
}
.toc-judul {
  flex: 1;
  line-height: 1.35;
}
.cocok {
  font-size: 10.5px;
  font-weight: 800;
  background: var(--brand);
  color: #fff;
  border-radius: 999px;
  padding: 1px 7px;
}
.kosong-kecil {
  font-size: 12.5px;
  color: var(--text-3);
  padding: 12px 6px;
}

/* Isi */
.konten {
  padding: 28px 34px 22px;
  min-width: 0;
}
.bab-kepala {
  padding-bottom: 18px;
  margin-bottom: 8px;
  border-bottom: 1px solid var(--border);
}
.bab-no {
  font-size: 11px;
  font-weight: 800;
  letter-spacing: 1px;
  text-transform: uppercase;
  color: var(--brand-aktif);
}
.bab-judul {
  font-size: 26px;
  font-weight: 800;
  letter-spacing: -0.4px;
  margin-top: 4px;
}
.bab-ringkas {
  margin-top: 8px;
  color: var(--text-2);
  font-size: 14.5px;
  line-height: 1.6;
  max-width: 75ch;
}
.bab-aksi {
  margin-top: 10px;
}
.sub-nav {
  margin-top: 12px;
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
}
.sub-chip {
  border: 1px solid var(--border-kuat);
  background: var(--surface);
  border-radius: 999px;
  padding: 4px 11px;
  font-size: 12px;
  font-weight: 600;
  color: var(--text-2);
}
.sub-chip:hover {
  border-color: var(--brand);
  color: var(--brand-aktif);
  background: var(--brand-light);
}

/* Prosa markdown (konten v-html → :deep) */
.prosa {
  font-size: 14.5px;
  line-height: 1.7;
  color: var(--text);
}
.prosa :deep(h2) {
  font-size: 19px;
  margin: 30px 0 10px;
  padding-top: 6px;
  scroll-margin-top: 90px;
}
.prosa :deep(h3) {
  font-size: 16px;
  margin: 22px 0 8px;
  scroll-margin-top: 90px;
}
.prosa :deep(p) {
  margin: 10px 0;
}
.prosa :deep(ul),
.prosa :deep(ol) {
  margin: 10px 0;
  padding-left: 22px;
}
.prosa :deep(li) {
  margin: 4px 0;
}
.prosa :deep(li input[type='checkbox']) {
  margin-right: 6px;
  accent-color: var(--brand);
}
.prosa :deep(strong) {
  color: var(--text);
}
.prosa :deep(code) {
  font-family: ui-monospace, 'SFMono-Regular', Menlo, monospace;
  font-size: 12.5px;
  background: var(--brand-light);
  color: var(--brand-gelap);
  padding: 1px 6px;
  border-radius: 5px;
}
.prosa :deep(.kode-blok) {
  position: relative;
  margin: 14px 0;
}
.prosa :deep(pre) {
  background: var(--kode-bg);
  border: 1px solid var(--border-kuat);
  color: #f3e6e2;
  border-radius: var(--radius-sm);
  padding: 16px 18px;
  overflow-x: auto;
  font-size: 12.5px;
  line-height: 1.6;
}
.prosa :deep(pre code) {
  background: transparent;
  color: inherit;
  padding: 0;
  font-size: inherit;
}
.prosa :deep(.salin) {
  position: absolute;
  top: 8px;
  right: 8px;
  border: 1px solid rgba(255, 255, 255, 0.18);
  background: rgba(255, 255, 255, 0.08);
  color: #f3e6e2;
  font-size: 11.5px;
  font-weight: 600;
  padding: 3px 9px;
  border-radius: 6px;
}
.prosa :deep(.salin:hover) {
  background: rgba(242, 105, 94, 0.35);
}
.prosa :deep(.kode-lang) {
  position: absolute;
  top: 10px;
  right: 72px;
  font-size: 10.5px;
  text-transform: uppercase;
  letter-spacing: 0.6px;
  color: rgba(243, 230, 226, 0.5);
}
.prosa :deep(blockquote) {
  margin: 14px 0;
  padding: 10px 16px;
  border-left: 4px solid var(--brand);
  background: var(--brand-light);
  border-radius: 0 var(--radius-sm) var(--radius-sm) 0;
  color: var(--text);
}
.prosa :deep(blockquote p) {
  margin: 4px 0;
}
.prosa :deep(table) {
  width: 100%;
  border-collapse: collapse;
  margin: 14px 0;
  font-size: 13px;
  display: block;
  overflow-x: auto;
}
.prosa :deep(th) {
  text-align: left;
  font-size: 11px;
  font-weight: 700;
  letter-spacing: 0.4px;
  text-transform: uppercase;
  color: var(--text-3);
  padding: 8px 12px;
  border-bottom: 1px solid var(--border-kuat);
  white-space: nowrap;
}
.prosa :deep(td) {
  padding: 8px 12px;
  border-bottom: 1px solid var(--border);
  vertical-align: top;
}
.prosa :deep(tr:hover td) {
  background: var(--bg);
}
.prosa :deep(hr) {
  border: none;
  border-top: 1px solid var(--border);
  margin: 24px 0;
}

/* Navigasi bab */
.navbab {
  display: flex;
  justify-content: space-between;
  gap: 12px;
  margin-top: 30px;
  padding-top: 18px;
  border-top: 1px solid var(--border);
}
.nav-btn {
  display: flex;
  align-items: center;
  gap: 10px;
  border: 1px solid var(--border);
  background: var(--surface);
  border-radius: var(--radius-sm);
  padding: 10px 14px;
  color: var(--text);
  font-weight: 700;
  font-size: 13.5px;
  text-align: left;
  max-width: 48%;
}
.nav-btn.kanan {
  text-align: right;
  margin-left: auto;
}
.nav-btn small {
  display: block;
  font-size: 11px;
  font-weight: 600;
  color: var(--text-3);
}
.nav-btn:hover {
  border-color: var(--brand);
  color: var(--brand-aktif);
}

@media (max-width: 980px) {
  .tata {
    grid-template-columns: 1fr;
  }
  .toc {
    position: static;
    max-height: 320px;
  }
  .konten {
    padding: 22px 18px;
  }
}
</style>
