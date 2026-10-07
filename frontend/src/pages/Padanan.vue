<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { useRoute } from 'vue-router'
import { Search, Pencil, Trash2, Check, X, Lightbulb } from 'lucide-vue-next'
import { api, GalatApi } from '@/lib/api'
import { angka } from '@/lib/format'
import PageHeader from '@/components/PageHeader.vue'

interface Row {
  kode: string
  deskripsi: string
  episode: number
  jumlah_cp: number
  padanan: string
  padanan_deskripsi: string
  status: string // SUDAH | BELUM | RAGU
  ditetapkan_oleh: string
  ditetapkan_pada: string
}
interface MasterItem { kode: string; deskripsi: string }
interface Resp {
  ringkasan: Record<string, number>
  boleh_ubah: boolean
  master: MasterItem[]
  data: Row[]
}

const route = useRoute()
const snomed = computed(() => route.name === 'snomed')
const basis = computed(() => (snomed.value ? 'padanan/snomed' : 'padanan/kptl'))
const labelPadanan = computed(() => (snomed.value ? 'SNOMED-CT' : 'KPTL'))
const labelKode = computed(() => (snomed.value ? 'ICD-10' : 'ICD-9-CM'))

const resp = ref<Resp | null>(null)
const galat = ref('')
const cari = ref('')
const filterStatus = ref('')

// editor baris
const editKode = ref('')
const pilih = ref('')
const usulan = ref<MasterItem[]>([])
const pesan = ref('')

async function muat() {
  galat.value = ''
  try {
    resp.value = await api<Resp>(basis.value)
  } catch (e) {
    galat.value = e instanceof GalatApi ? e.message : 'Gagal memuat data.'
  }
}
watch(basis, muat, { immediate: true })

const hasil = computed(() => {
  if (!resp.value) return []
  const q = cari.value.trim().toLowerCase()
  return resp.value.data.filter((r) => {
    const cocok = !q || r.kode.toLowerCase().includes(q) || r.deskripsi.toLowerCase().includes(q)
    const cocokStatus = !filterStatus.value || r.status === filterStatus.value
    return cocok && cocokStatus
  })
})

async function mulaiEdit(r: Row) {
  editKode.value = r.kode
  pilih.value = r.padanan || ''
  pesan.value = ''
  usulan.value = []
  try {
    const u = await api<{ data: MasterItem[] }>(`${basis.value}/${encodeURIComponent(r.kode)}/usulan`)
    usulan.value = u.data
    if (!pilih.value && u.data.length) pilih.value = u.data[0].kode
  } catch {
    /* abaikan */
  }
}

async function simpan(r: Row) {
  if (!pilih.value) return
  pesan.value = ''
  const body = snomed.value ? { icd10: r.kode, snomed: pilih.value } : { icd9cm: r.kode, kptl: pilih.value }
  try {
    resp.value = await api<Resp>(basis.value, { method: 'POST', body: JSON.stringify(body) })
    editKode.value = ''
  } catch (e) {
    pesan.value = e instanceof GalatApi ? e.message : 'Gagal menyimpan.'
  }
}

async function hapus(r: Row) {
  try {
    resp.value = await api<Resp>(`${basis.value}/${encodeURIComponent(r.kode)}`, { method: 'DELETE' })
  } catch (e) {
    pesan.value = e instanceof GalatApi ? e.message : 'Gagal menghapus.'
  }
}

const statusGaya: Record<string, { bg: string; fg: string; label: string }> = {
  SUDAH: { bg: 'var(--hijau-bg)', fg: 'var(--hijau)', label: 'Sudah' },
  BELUM: { bg: 'var(--kuning-bg)', fg: 'var(--kuning)', label: 'Belum' },
  RAGU: { bg: 'var(--ungu-bg)', fg: 'var(--ungu)', label: 'Ragu' },
}

const statusFilter = computed(() =>
  snomed.value
    ? [{ v: '', l: 'Semua' }, { v: 'BELUM', l: 'Belum' }, { v: 'RAGU', l: 'Ragu' }, { v: 'SUDAH', l: 'Sudah' }]
    : [{ v: '', l: 'Semua' }, { v: 'BELUM', l: 'Belum' }, { v: 'SUDAH', l: 'Sudah' }],
)
</script>

<template>
  <div>
    <PageHeader
      :judul="`Padanan ${labelPadanan}`"
      demo
      :sub="snomed
        ? 'Pemetaan diagnosis ICD-10 ke konsep SNOMED-CT untuk interoperabilitas rekam medis.'
        : 'Pemetaan prosedur ICD-9-CM ke Kode Prosedur Tarif Lokal (KPTL) agar rencana tindakan selaras tarif.'"
    />

    <div v-if="galat" class="galat-box">{{ galat }}</div>

    <template v-if="resp">
      <!-- Ringkasan -->
      <div class="ringkas">
        <div class="rk card">
          <p class="rk-n angka">{{ angka(resp.ringkasan[snomed ? 'jumlah_diagnosis' : 'jumlah_prosedur']) }}</p>
          <p class="rk-l">{{ snomed ? 'Diagnosis' : 'Prosedur' }}</p>
        </div>
        <div class="rk card">
          <p class="rk-n angka" style="color: var(--hijau)">{{ angka(resp.ringkasan.sudah) }}</p>
          <p class="rk-l">Sudah dipetakan</p>
        </div>
        <div v-if="snomed" class="rk card">
          <p class="rk-n angka" style="color: var(--ungu)">{{ angka(resp.ringkasan.ragu) }}</p>
          <p class="rk-l">Ragu (perlu konfirmasi)</p>
        </div>
        <div class="rk card">
          <p class="rk-n angka" style="color: var(--kuning)">{{ angka(resp.ringkasan.belum) }}</p>
          <p class="rk-l">Belum dipetakan</p>
        </div>
        <div v-if="!snomed" class="rk card">
          <p class="rk-n angka" style="color: var(--brand-aktif)">{{ resp.ringkasan.cakupan_episode }}%</p>
          <p class="rk-l">Cakupan episode</p>
        </div>
      </div>

      <!-- Filter -->
      <div class="card filter-bar">
        <div class="cari">
          <Search :size="17" class="cari-ikon" />
          <input v-model="cari" :placeholder="`Cari ${labelKode} atau deskripsi…`" />
        </div>
        <div class="seg">
          <button v-for="f in statusFilter" :key="f.v" class="seg-btn" :class="{ aktif: filterStatus === f.v }" @click="filterStatus = f.v">{{ f.l }}</button>
        </div>
      </div>

      <p v-if="pesan" class="pesan-edit">{{ pesan }}</p>

      <!-- Tabel -->
      <div class="card panel">
        <div class="tabel-bungkus">
          <table class="tabel">
            <thead>
              <tr>
                <th>{{ labelKode }}</th>
                <th>Deskripsi</th>
                <th class="kanan">Episode</th>
                <th>Status</th>
                <th>Padanan {{ labelPadanan }}</th>
                <th v-if="resp.boleh_ubah"></th>
              </tr>
            </thead>
            <tbody>
              <template v-for="r in hasil" :key="r.kode">
                <tr>
                  <td><span class="mono kode">{{ r.kode }}</span></td>
                  <td class="desk">{{ r.deskripsi }}</td>
                  <td class="kanan angka">{{ angka(r.episode) }}</td>
                  <td>
                    <span class="badge" :style="{ background: statusGaya[r.status].bg, color: statusGaya[r.status].fg }">{{ statusGaya[r.status].label }}</span>
                  </td>
                  <td>
                    <template v-if="r.padanan">
                      <span class="mono pad-kode">{{ r.padanan }}</span>
                      <span class="pad-desk">{{ r.padanan_deskripsi }}</span>
                      <span v-if="r.ditetapkan_oleh" class="pad-oleh">· {{ r.ditetapkan_oleh }}, {{ r.ditetapkan_pada }}</span>
                    </template>
                    <span v-else class="dimmed">—</span>
                  </td>
                  <td v-if="resp.boleh_ubah" class="kanan nowrap">
                    <button v-if="editKode !== r.kode" class="btn btn-outline btn-xs" @click="mulaiEdit(r)">
                      <Pencil :size="13" /> {{ r.padanan ? 'Ubah' : 'Tetapkan' }}
                    </button>
                    <button v-if="r.padanan && editKode !== r.kode" class="hapus-btn" title="Cabut" @click="hapus(r)"><Trash2 :size="14" /></button>
                  </td>
                </tr>
                <!-- editor -->
                <tr v-if="editKode === r.kode" class="editor-row">
                  <td :colspan="resp.boleh_ubah ? 6 : 5">
                    <div class="editor">
                      <div v-if="usulan.length" class="usulan">
                        <Lightbulb :size="14" /> <span class="usulan-l">Usulan:</span>
                        <button v-for="u in usulan" :key="u.kode" class="usulan-chip" :class="{ aktif: pilih === u.kode }" @click="pilih = u.kode" :title="u.deskripsi">{{ u.kode }}</button>
                      </div>
                      <div class="editor-aksi">
                        <select v-model="pilih" class="inp">
                          <option value="" disabled>Pilih {{ labelPadanan }}…</option>
                          <option v-for="m in resp.master" :key="m.kode" :value="m.kode">{{ m.kode }} — {{ m.deskripsi }}</option>
                        </select>
                        <button class="btn btn-ghost btn-xs" @click="editKode = ''"><X :size="14" /> Batal</button>
                        <button class="btn btn-brand btn-xs" :disabled="!pilih" @click="simpan(r)"><Check :size="14" /> Simpan</button>
                      </div>
                    </div>
                  </td>
                </tr>
              </template>
              <tr v-if="!hasil.length"><td colspan="6" class="kosong">Tidak ada data yang cocok.</td></tr>
            </tbody>
          </table>
        </div>
      </div>
    </template>
  </div>
</template>

<style scoped>
.galat-box {
  padding: 14px 16px;
  border-radius: var(--radius);
  background: var(--merah-bg);
  color: var(--merah);
  font-weight: 600;
}
.ringkas {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(150px, 1fr));
  gap: 14px;
  margin-bottom: 16px;
}
.rk {
  padding: 16px 18px;
}
.rk-n {
  font-family: var(--font-judul);
  font-size: 26px;
  font-weight: 800;
}
.rk-l {
  font-size: 12.5px;
  color: var(--text-2);
  margin-top: 2px;
}
.filter-bar {
  display: flex;
  gap: 12px;
  padding: 12px;
  margin-bottom: 16px;
  flex-wrap: wrap;
  align-items: center;
}
.cari {
  flex: 1;
  min-width: 200px;
  display: flex;
  align-items: center;
  gap: 8px;
  height: 40px;
  padding: 0 12px;
  background: var(--bg);
  border: 1px solid var(--border);
  border-radius: var(--radius-sm);
}
.cari-ikon {
  color: var(--text-3);
}
.cari input {
  flex: 1;
  border: none;
  outline: none;
  background: transparent;
  font-size: 14px;
}
.seg {
  display: flex;
  gap: 4px;
  background: var(--bg);
  border: 1px solid var(--border);
  border-radius: var(--radius-sm);
  padding: 3px;
}
.seg-btn {
  border: none;
  background: transparent;
  padding: 6px 12px;
  border-radius: 7px;
  font-size: 13px;
  font-weight: 600;
  color: var(--text-2);
}
.seg-btn.aktif {
  background: var(--surface);
  color: var(--brand-aktif);
  box-shadow: var(--shadow);
}
.pesan-edit {
  color: var(--merah);
  font-size: 13px;
  margin-bottom: 10px;
}
.panel {
  padding: 8px 0;
}
.tabel-bungkus {
  overflow-x: auto;
}
.tabel {
  width: 100%;
  border-collapse: collapse;
  font-size: 13.5px;
}
.tabel th {
  text-align: left;
  font-size: 10.5px;
  font-weight: 700;
  letter-spacing: 0.5px;
  text-transform: uppercase;
  color: var(--text-3);
  padding: 10px 14px;
  border-bottom: 1px solid var(--border);
}
.tabel th:first-child,
.tabel td:first-child {
  padding-left: 22px;
}
.tabel th:last-child,
.tabel td:last-child {
  padding-right: 22px;
}
.tabel td {
  padding: 11px 14px;
  border-bottom: 1px solid var(--border);
}
.tabel tbody tr:hover {
  background: var(--bg);
}
.editor-row:hover {
  background: transparent !important;
}
.kanan {
  text-align: right;
}
.nowrap {
  white-space: nowrap;
}
.kode {
  font-size: 12.5px;
  font-weight: 700;
  color: var(--brand-aktif);
}
.desk {
  max-width: 340px;
}
.pad-kode {
  font-size: 12px;
  font-weight: 700;
  color: var(--text);
}
.pad-desk {
  color: var(--text-2);
  margin-left: 6px;
}
.pad-oleh {
  color: var(--text-3);
  font-size: 12px;
  margin-left: 4px;
}
.btn-xs {
  height: 30px;
  padding: 0 10px;
  font-size: 12.5px;
}
.hapus-btn {
  border: none;
  background: transparent;
  color: var(--text-3);
  width: 30px;
  height: 30px;
  border-radius: 8px;
  margin-left: 4px;
}
.hapus-btn:hover {
  background: var(--merah-bg);
  color: var(--merah);
}
.editor {
  background: var(--bg);
  border-radius: var(--radius-sm);
  padding: 12px 14px;
  display: flex;
  flex-direction: column;
  gap: 10px;
}
.usulan {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
  color: var(--text-2);
  font-size: 12.5px;
}
.usulan-l {
  color: var(--text-3);
}
.usulan-chip {
  border: 1px solid var(--border-kuat);
  background: var(--surface);
  border-radius: 999px;
  padding: 3px 10px;
  font-size: 12px;
  font-weight: 700;
  color: var(--text-2);
}
.usulan-chip.aktif {
  border-color: var(--brand);
  color: var(--brand-aktif);
  background: var(--brand-light);
}
.editor-aksi {
  display: flex;
  gap: 8px;
  align-items: center;
}
.inp {
  flex: 1;
  height: 38px;
  padding: 0 10px;
  border: 1px solid var(--border-kuat);
  border-radius: 8px;
  background: var(--surface);
  font-size: 13.5px;
  color: var(--text);
}
.kosong {
  text-align: center;
  padding: 36px;
  color: var(--text-3);
}
</style>
