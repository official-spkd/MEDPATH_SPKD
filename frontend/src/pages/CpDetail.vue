<script setup lang="ts">
import { ref, computed, watch, inject } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import {
  ArrowLeft,
  Check,
  X,
  CircleCheck,
  CircleAlert,
  BookOpenText,
  Pill,
  ListChecks,
  History,
  LayoutList,
  Stamp,
  Undo2,
  Plus,
  Trash2,
} from 'lucide-vue-next'
import { api, GalatApi } from '@/lib/api'
import { angka } from '@/lib/format'
import StatusPill from '@/components/StatusPill.vue'

interface Aksi { aksi: string; ke: string; label: string }
interface Syarat { kode: string; judul: string; menghambat: boolean; terpenuhi: boolean; penanggung_jawab: string; kekurangan: string[] }
interface Detail {
  kode: string
  nama: string
  kelompok_diagnosis: string
  status: string
  status_label: string
  status_sejak: string
  jumlah_episode: number
  los_median: number
  severity: { severity: string; episode: number; los_median: number }[]
  sub_cp: { icd10: string; diagnosa: string; jumlah_episode: number; porsi: string; ada_acuan: boolean }[]
  acuan: { id: number; dokumen: string; sumber: string; tahun: number; catatan: string }[]
  kriteria: { id: number; jenis: string; uraian: string }[]
  rencana: { id: number; jenis: string; severity: string; hari: number; sifat: string; kode: string; nama: string; dosis: string; rute: string; frekuensi: string }[]
  syarat: Syarat[]
  aksi_tersedia: Aksi[]
  menunggu_peran_list: string[]
  boleh_ubah: { acuan: boolean; rencana: boolean; kriteria: boolean }
}
interface RiwayatItem { id: number; dari_status: string; ke_status: string; aksi: string; peran: string; user: string; alasan: string; waktu: string }
interface DokumenRingkas { id: number; nama: string; tahun: number }

const route = useRoute()
const router = useRouter()
const refreshAntrean = inject<() => void>('refreshAntrean', () => {})

const d = ref<Detail | null>(null)
const riwayat = ref<RiwayatItem[]>([])
const galat = ref('')
const tab = ref<'ringkasan' | 'acuan' | 'rencana' | 'kriteria' | 'riwayat'>('ringkasan')

// Konfirmasi aksi
const aksiAktif = ref<Aksi | null>(null)
const alasan = ref('')
const proses = ref(false)
const pesanAksi = ref('')

const kode = computed(() => route.params.kode as string)

async function muat() {
  galat.value = ''
  try {
    d.value = await api<Detail>(`cp/${kode.value}`)
    const r = await api<{ data: RiwayatItem[] }>(`cp/${kode.value}/riwayat`)
    riwayat.value = r.data
  } catch (e) {
    galat.value = e instanceof GalatApi ? e.message : 'Gagal memuat detail CP.'
  }
}
// Jalan saat mount dan tiap kode berubah (lebih andal dari onMounted untuk async component).
watch(kode, muat, { immediate: true })

// --- Edit isi klinis ---
const dokumenList = ref<DokumenRingkas[]>([])
const pesanEdit = ref('')

async function muatDokumen() {
  if (dokumenList.value.length) return
  try {
    const r = await api<{ data: DokumenRingkas[] }>('dokumen-panduan')
    dokumenList.value = r.data
  } catch {
    /* abaikan */
  }
}

// Form tambah
const formAcuan = ref({ dokumen_id: 0, catatan: '' })
const tambahAcuanBuka = ref(false)
const formKriteria = ref({ jenis: 'INKLUSI', uraian: '' })
const tambahKriteriaBuka = ref(false)
const formRencana = ref({ jenis: 'OBAT', severity: 'I', nama: '', kode: '', dosis: '', rute: '', frekuensi: '', hari: 1, sifat: 'WAJIB' })
const tambahRencanaBuka = ref(false)

async function kirimEdit(fn: () => Promise<Detail>, reset?: () => void) {
  pesanEdit.value = ''
  try {
    d.value = await fn()
    reset?.()
  } catch (e) {
    pesanEdit.value = e instanceof GalatApi ? e.message : 'Gagal menyimpan.'
  }
}

const tambahAcuan = () =>
  kirimEdit(
    () => api<Detail>(`cp/${kode.value}/acuan`, { method: 'POST', body: JSON.stringify(formAcuan.value) }),
    () => {
      formAcuan.value = { dokumen_id: 0, catatan: '' }
      tambahAcuanBuka.value = false
    },
  )
const hapusAcuan = (id: number) => kirimEdit(() => api<Detail>(`cp/${kode.value}/acuan/${id}`, { method: 'DELETE' }))

const tambahKriteria = () =>
  kirimEdit(
    () => api<Detail>(`cp/${kode.value}/kriteria`, { method: 'POST', body: JSON.stringify(formKriteria.value) }),
    () => {
      formKriteria.value = { jenis: 'INKLUSI', uraian: '' }
      tambahKriteriaBuka.value = false
    },
  )
const hapusKriteria = (id: number) => kirimEdit(() => api<Detail>(`cp/${kode.value}/kriteria/${id}`, { method: 'DELETE' }))

const tambahRencana = () =>
  kirimEdit(
    () => api<Detail>(`cp/${kode.value}/rencana`, { method: 'POST', body: JSON.stringify(formRencana.value) }),
    () => {
      formRencana.value = { jenis: 'OBAT', severity: 'I', nama: '', kode: '', dosis: '', rute: '', frekuensi: '', hari: 1, sifat: 'WAJIB' }
      tambahRencanaBuka.value = false
    },
  )
const hapusRencana = (id: number) => kirimEdit(() => api<Detail>(`cp/${kode.value}/rencana/${id}`, { method: 'DELETE' }))

watch(tambahAcuanBuka, (v) => v && muatDokumen())

function mulaiAksi(a: Aksi) {
  aksiAktif.value = a
  alasan.value = ''
  pesanAksi.value = ''
}

async function jalankanAksi() {
  if (!aksiAktif.value) return
  proses.value = true
  pesanAksi.value = ''
  try {
    await api(`cp/${kode.value}/transisi`, {
      method: 'POST',
      body: JSON.stringify({ aksi: aksiAktif.value.aksi, alasan: alasan.value }),
    })
    aksiAktif.value = null
    await muat()
    refreshAntrean()
  } catch (e) {
    pesanAksi.value = e instanceof GalatApi ? e.message : 'Gagal menjalankan aksi.'
  } finally {
    proses.value = false
  }
}

const syaratTerpenuhi = computed(() => (d.value?.syarat?.filter((s) => s.terpenuhi).length) ?? 0)
const ikonAksi: Record<string, any> = { AJUKAN: Stamp, SETUJUI: Check, SAHKAN: CircleCheck, KEMBALIKAN: Undo2 }
const warnaJenis: Record<string, string> = { OBAT: 'brand', LAB: 'info', PROSEDUR: 'ungu', TINDAKAN: 'kuning' }

const tabs = [
  { id: 'ringkasan', label: 'Ringkasan', icon: LayoutList },
  { id: 'acuan', label: 'Acuan', icon: BookOpenText },
  { id: 'rencana', label: 'Rencana Klinis', icon: Pill },
  { id: 'kriteria', label: 'Kriteria', icon: ListChecks },
  { id: 'riwayat', label: 'Riwayat', icon: History },
] as const
</script>

<template>
  <div v-if="galat" class="galat-box"><CircleAlert :size="18" /> {{ galat }}</div>

  <div v-else-if="!d" class="memuat">Memuat detail CP…</div>

  <div v-else class="detail">
    <!-- Header -->
    <div class="head">
      <button class="btn btn-ghost btn-sm kembali" @click="router.push('/library')"><ArrowLeft :size="16" /> CP Library</button>
      <div class="head-main">
        <div>
          <div class="judul-baris">
            <span class="mono kode-besar">{{ d.kode }}</span>
            <StatusPill :status="d.status" />
          </div>
          <h1>{{ d.nama }}</h1>
          <p class="sub">{{ d.kelompok_diagnosis }} · {{ angka(d.jumlah_episode) }} episode · LOS median {{ d.los_median }} hari</p>
        </div>
      </div>
    </div>

    <div class="kolom">
      <!-- Kiri: konten tab -->
      <div class="konten">
        <div class="tabbar">
          <button
            v-for="t in tabs"
            :key="t.id"
            class="tab"
            :class="{ aktif: tab === t.id }"
            @click="tab = t.id as any"
          >
            <component :is="t.icon" :size="16" /> {{ t.label }}
            <span v-if="t.id === 'riwayat'" class="tab-badge angka">{{ riwayat.length }}</span>
          </button>
        </div>

        <!-- Ringkasan -->
        <div v-if="tab === 'ringkasan'" class="card panel">
          <p class="panel-judul">Statistik per severity</p>
          <div class="sev-grid">
            <div v-for="s in d.severity" :key="s.severity" class="sev">
              <p class="sev-label">Severity {{ s.severity }}</p>
              <p class="sev-ep angka">{{ angka(s.episode) }}</p>
              <p class="sev-los">episode · LOS {{ s.los_median }} hari</p>
            </div>
          </div>

          <p class="panel-judul mt">Sub-CP (diagnosis)</p>
          <table class="tabel">
            <thead><tr><th>ICD-10</th><th>Diagnosa</th><th class="kanan">Episode</th><th class="kanan">Acuan</th></tr></thead>
            <tbody>
              <tr v-for="s in d.sub_cp" :key="s.icd10">
                <td><span class="mono kode">{{ s.icd10 }}</span></td>
                <td>{{ s.diagnosa }}</td>
                <td class="kanan angka">{{ angka(s.jumlah_episode) }}</td>
                <td class="kanan">
                  <CircleCheck v-if="s.ada_acuan" :size="16" class="ok-ikon" />
                  <span v-else class="badge kurang">belum</span>
                </td>
              </tr>
            </tbody>
          </table>
        </div>

        <!-- Acuan -->
        <div v-else-if="tab === 'acuan'" class="card panel">
          <div class="panel-head">
            <p class="panel-judul">Acuan klinis ({{ d.acuan?.length }})</p>
            <button v-if="d.boleh_ubah?.acuan" class="btn btn-brand btn-sm" @click="tambahAcuanBuka = !tambahAcuanBuka">
              <Plus :size="15" /> Tambah acuan
            </button>
          </div>

          <div v-if="tambahAcuanBuka" class="form-tambah">
            <select v-model.number="formAcuan.dokumen_id" class="inp">
              <option :value="0" disabled>Pilih dokumen panduan…</option>
              <option v-for="dk in dokumenList" :key="dk.id" :value="dk.id">{{ dk.nama }} ({{ dk.tahun }})</option>
            </select>
            <input v-model="formAcuan.catatan" class="inp" placeholder="Catatan (opsional)" />
            <div class="form-aksi">
              <button class="btn btn-ghost btn-sm" @click="tambahAcuanBuka = false">Batal</button>
              <button class="btn btn-brand btn-sm" :disabled="!formAcuan.dokumen_id" @click="tambahAcuan">Simpan</button>
            </div>
          </div>
          <p v-if="pesanEdit" class="pesan-edit"><CircleAlert :size="14" /> {{ pesanEdit }}</p>

          <div v-if="!d.acuan?.length" class="kosong-kecil">Belum ada acuan ditetapkan.</div>
          <ul v-else class="acuan-list">
            <li v-for="a in d.acuan" :key="a.id">
              <span class="acuan-ikon"><BookOpenText :size="17" /></span>
              <div class="acuan-isi">
                <p class="acuan-nama">{{ a.dokumen }}</p>
                <p class="acuan-meta">{{ a.sumber }} · {{ a.tahun }}</p>
                <p v-if="a.catatan" class="acuan-catatan">{{ a.catatan }}</p>
              </div>
              <button v-if="d.boleh_ubah?.acuan" class="hapus-btn" title="Hapus acuan" @click="hapusAcuan(a.id)">
                <Trash2 :size="15" />
              </button>
            </li>
          </ul>
        </div>

        <!-- Rencana -->
        <div v-else-if="tab === 'rencana'" class="card panel">
          <div class="panel-head">
            <p class="panel-judul">Rencana isi klinis ({{ d.rencana?.length }})</p>
            <button v-if="d.boleh_ubah?.rencana" class="btn btn-brand btn-sm" @click="tambahRencanaBuka = !tambahRencanaBuka">
              <Plus :size="15" /> Tambah rencana
            </button>
          </div>

          <div v-if="tambahRencanaBuka" class="form-tambah form-grid">
            <select v-model="formRencana.jenis" class="inp"><option>OBAT</option><option>LAB</option><option>PROSEDUR</option><option>TINDAKAN</option></select>
            <select v-model="formRencana.severity" class="inp"><option>I</option><option>II</option><option>III</option></select>
            <input v-model="formRencana.nama" class="inp lebar" placeholder="Nama (mis. Seftriakson)" />
            <input v-model="formRencana.kode" class="inp" placeholder="Kode (KFA/LOINC)" />
            <input v-model="formRencana.dosis" class="inp" placeholder="Dosis" />
            <input v-model="formRencana.rute" class="inp" placeholder="Rute" />
            <input v-model="formRencana.frekuensi" class="inp" placeholder="Frekuensi" />
            <input v-model.number="formRencana.hari" type="number" min="1" class="inp sempit" placeholder="Hari" />
            <select v-model="formRencana.sifat" class="inp"><option>WAJIB</option><option>KONDISIONAL</option></select>
            <div class="form-aksi span">
              <button class="btn btn-ghost btn-sm" @click="tambahRencanaBuka = false">Batal</button>
              <button class="btn btn-brand btn-sm" :disabled="!formRencana.nama" @click="tambahRencana">Simpan</button>
            </div>
          </div>
          <p v-if="pesanEdit" class="pesan-edit"><CircleAlert :size="14" /> {{ pesanEdit }}</p>

          <div v-if="!d.rencana?.length" class="kosong-kecil">Belum ada rencana klinis.</div>
          <table v-else class="tabel">
            <thead><tr><th>Sev</th><th>Jenis</th><th>Nama</th><th>Dosis/Rute</th><th>Frekuensi</th><th class="kanan">Hari</th><th>Sifat</th><th v-if="d.boleh_ubah?.rencana"></th></tr></thead>
            <tbody>
              <tr v-for="r in d.rencana" :key="r.id">
                <td><span class="mono">{{ r.severity }}</span></td>
                <td><span class="badge" :class="`s-${warnaJenis[r.jenis]}`">{{ r.jenis }}</span></td>
                <td class="nama">{{ r.nama }}<br /><span class="mono kode-kecil">{{ r.kode }}</span></td>
                <td class="muted">{{ r.dosis }}<template v-if="r.rute !== '—'"> · {{ r.rute }}</template></td>
                <td class="muted">{{ r.frekuensi }}</td>
                <td class="kanan angka">{{ r.hari }}</td>
                <td><span class="badge" :class="r.sifat === 'WAJIB' ? 'ok' : 'kurang'">{{ r.sifat }}</span></td>
                <td v-if="d.boleh_ubah?.rencana" class="kanan">
                  <button class="hapus-btn" title="Hapus" @click="hapusRencana(r.id)"><Trash2 :size="15" /></button>
                </td>
              </tr>
            </tbody>
          </table>
        </div>

        <!-- Kriteria -->
        <div v-else-if="tab === 'kriteria'" class="card panel">
          <div class="panel-head">
            <p class="panel-judul">Kriteria inklusi & eksklusi</p>
            <button v-if="d.boleh_ubah?.kriteria" class="btn btn-brand btn-sm" @click="tambahKriteriaBuka = !tambahKriteriaBuka">
              <Plus :size="15" /> Tambah kriteria
            </button>
          </div>

          <div v-if="tambahKriteriaBuka" class="form-tambah">
            <div class="form-baris">
              <select v-model="formKriteria.jenis" class="inp"><option value="INKLUSI">Inklusi</option><option value="EKSKLUSI">Eksklusi</option></select>
              <input v-model="formKriteria.uraian" class="inp lebar" placeholder="Uraian kriteria…" />
            </div>
            <div class="form-aksi">
              <button class="btn btn-ghost btn-sm" @click="tambahKriteriaBuka = false">Batal</button>
              <button class="btn btn-brand btn-sm" :disabled="!formKriteria.uraian" @click="tambahKriteria">Simpan</button>
            </div>
          </div>
          <p v-if="pesanEdit" class="pesan-edit"><CircleAlert :size="14" /> {{ pesanEdit }}</p>

          <div v-if="!d.kriteria?.length" class="kosong-kecil">Belum ada kriteria.</div>
          <ul v-else class="kriteria-list">
            <li v-for="k in d.kriteria" :key="k.id">
              <span class="badge" :class="k.jenis === 'INKLUSI' ? 'ok' : 'kurang'">{{ k.jenis }}</span>
              <span class="kriteria-uraian">{{ k.uraian }}</span>
              <button v-if="d.boleh_ubah?.kriteria" class="hapus-btn" title="Hapus" @click="hapusKriteria(k.id)"><Trash2 :size="15" /></button>
            </li>
          </ul>
        </div>

        <!-- Riwayat -->
        <div v-else class="card panel">
          <p class="panel-judul">Riwayat pengesahan</p>
          <div v-if="!riwayat.length" class="kosong-kecil">Belum ada riwayat.</div>
          <ol v-else class="timeline">
            <li v-for="r in riwayat" :key="r.id">
              <span class="tl-dot" />
              <div class="tl-isi">
                <p class="tl-aksi">{{ r.aksi }} <StatusPill :status="r.ke_status" /></p>
                <p class="tl-meta">{{ r.peran }} · {{ r.waktu }}</p>
                <p v-if="r.alasan" class="tl-alasan">"{{ r.alasan }}"</p>
              </div>
            </li>
          </ol>
        </div>
      </div>

      <!-- Kanan: pengesahan -->
      <aside class="samping">
        <div class="card panel">
          <p class="panel-judul">Pengesahan</p>

          <div v-if="d.menunggu_peran_list?.length" class="menunggu">
            Menunggu tindakan
            <span v-for="p in d.menunggu_peran_list" :key="p" class="badge peran">{{ p }}</span>
          </div>
          <p v-else-if="d.status === 'AKTIF'" class="aktif-note"><CircleCheck :size="16" /> CP sudah aktif & memandu pelayanan.</p>

          <!-- Aksi -->
          <div v-if="d.aksi_tersedia?.length" class="aksi-wrap">
            <template v-if="!aksiAktif">
              <button
                v-for="a in d.aksi_tersedia"
                :key="a.aksi"
                class="btn aksi-btn"
                :class="a.aksi === 'KEMBALIKAN' ? 'btn-outline' : 'btn-brand'"
                @click="mulaiAksi(a)"
              >
                <component :is="ikonAksi[a.aksi]" :size="16" /> {{ a.label }}
              </button>
            </template>
            <div v-else class="konfirmasi">
              <p class="konf-judul">{{ aksiAktif.label }}?</p>
              <textarea
                v-if="aksiAktif.aksi === 'KEMBALIKAN'"
                v-model="alasan"
                placeholder="Alasan pengembalian (wajib)…"
                rows="3"
              />
              <p v-if="pesanAksi" class="konf-galat"><CircleAlert :size="14" /> {{ pesanAksi }}</p>
              <div class="konf-aksi">
                <button class="btn btn-ghost btn-sm" @click="aksiAktif = null">Batal</button>
                <button
                  class="btn btn-brand btn-sm"
                  :disabled="proses || (aksiAktif.aksi === 'KEMBALIKAN' && !alasan.trim())"
                  @click="jalankanAksi"
                >
                  {{ proses ? 'Memproses…' : 'Konfirmasi' }}
                </button>
              </div>
            </div>
          </div>
          <p v-else class="no-aksi">Tidak ada aksi untuk peran Anda pada status ini.</p>
        </div>

        <!-- Syarat -->
        <div class="card panel">
          <div class="panel-head">
            <p class="panel-judul">Syarat kelengkapan</p>
            <span class="badge" :class="syaratTerpenuhi === d.syarat.length ? 'ok' : 'kurang'">{{ syaratTerpenuhi }}/{{ d.syarat.length }}</span>
          </div>
          <ul class="syarat-list">
            <li v-for="s in d.syarat" :key="s.kode" :class="{ belum: !s.terpenuhi }">
              <span class="syarat-ikon" :class="s.terpenuhi ? 'ya' : 'tidak'">
                <Check v-if="s.terpenuhi" :size="13" />
                <X v-else :size="13" />
              </span>
              <div>
                <p class="syarat-judul">
                  {{ s.judul }}
                  <span v-if="s.menghambat" class="badge mini-hambat">wajib</span>
                </p>
                <p class="syarat-pj">{{ s.penanggung_jawab }}</p>
                <ul v-if="s.kekurangan.length" class="kekurangan">
                  <li v-for="(k, i) in s.kekurangan" :key="i">{{ k }}</li>
                </ul>
              </div>
            </li>
          </ul>
        </div>
      </aside>
    </div>
  </div>
</template>

<style scoped>
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
.memuat {
  padding: 60px;
  text-align: center;
  color: var(--text-3);
}
.detail {
  display: flex;
  flex-direction: column;
  gap: 18px;
}
.kembali {
  margin-bottom: 10px;
}
.judul-baris {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 6px;
}
.kode-besar {
  font-size: 14px;
  font-weight: 700;
  color: var(--brand-aktif);
}
.head h1 {
  font-size: 24px;
  font-weight: 800;
  letter-spacing: -0.5px;
}
.sub {
  margin-top: 6px;
  color: var(--text-2);
  font-size: 14px;
}
.kolom {
  display: grid;
  grid-template-columns: 1fr 340px;
  gap: 18px;
  align-items: start;
}
.konten {
  min-width: 0;
}
.tabbar {
  display: flex;
  gap: 4px;
  margin-bottom: 14px;
  border-bottom: 1px solid var(--border);
  overflow-x: auto;
}
.tab {
  display: inline-flex;
  align-items: center;
  gap: 7px;
  padding: 10px 14px;
  border: none;
  background: transparent;
  color: var(--text-2);
  font-weight: 600;
  font-size: 13.5px;
  border-bottom: 2px solid transparent;
  white-space: nowrap;
}
.tab:hover {
  color: var(--text);
}
.tab.aktif {
  color: var(--brand-aktif);
  border-bottom-color: var(--brand);
}
.tab-badge {
  background: var(--brand-light);
  color: var(--brand-aktif);
  font-size: 11px;
  font-weight: 700;
  padding: 1px 7px;
  border-radius: 999px;
}
.panel {
  padding: 20px 22px;
}
.panel-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 14px;
}
.panel-judul {
  font-family: var(--font-judul);
  font-weight: 700;
  font-size: 15px;
}
.mt {
  margin-top: 22px;
}
.sev-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 12px;
  margin-top: 12px;
}
.sev {
  padding: 14px;
  border: 1px solid var(--border);
  border-radius: var(--radius-sm);
  text-align: center;
}
.sev-label {
  font-size: 12px;
  color: var(--text-3);
  font-weight: 700;
}
.sev-ep {
  font-family: var(--font-judul);
  font-size: 24px;
  font-weight: 800;
  margin: 4px 0;
}
.sev-los {
  font-size: 11.5px;
  color: var(--text-3);
}
.tabel {
  width: 100%;
  border-collapse: collapse;
  font-size: 13.5px;
  margin-top: 10px;
}
.tabel th {
  text-align: left;
  font-size: 10.5px;
  font-weight: 700;
  letter-spacing: 0.5px;
  text-transform: uppercase;
  color: var(--text-3);
  padding: 8px 10px;
  border-bottom: 1px solid var(--border);
}
.tabel td {
  padding: 10px;
  border-bottom: 1px solid var(--border);
  vertical-align: top;
}
.tabel tbody tr:last-child td {
  border-bottom: none;
}
.kanan {
  text-align: right;
}
.kode {
  font-size: 12px;
  font-weight: 700;
  color: var(--brand-aktif);
}
.kode-kecil {
  font-size: 11px;
  color: var(--text-3);
}
.nama {
  font-weight: 600;
}
.ok-ikon {
  color: var(--hijau);
}
.ok {
  background: var(--hijau-bg);
  color: var(--hijau);
}
.kurang {
  background: var(--kuning-bg);
  color: var(--kuning);
}
.s-brand { background: var(--brand-light); color: var(--brand-aktif); }
.s-info { background: var(--info-bg); color: var(--info); }
.s-ungu { background: var(--ungu-bg); color: var(--ungu); }
.s-kuning { background: var(--kuning-bg); color: var(--kuning); }
.kosong-kecil {
  padding: 24px;
  text-align: center;
  color: var(--text-3);
  font-size: 13px;
}
.acuan-list,
.kriteria-list {
  list-style: none;
  padding: 0;
  display: flex;
  flex-direction: column;
  gap: 10px;
}
.acuan-list li {
  display: flex;
  gap: 12px;
  padding: 12px;
  border: 1px solid var(--border);
  border-radius: var(--radius-sm);
}
.acuan-ikon {
  width: 36px;
  height: 36px;
  flex-shrink: 0;
  border-radius: 9px;
  display: grid;
  place-items: center;
  background: var(--brand-light);
  color: var(--brand-aktif);
}
.acuan-nama {
  font-weight: 600;
  font-size: 14px;
}
.acuan-meta {
  font-size: 12px;
  color: var(--text-3);
  margin-top: 2px;
}
.acuan-catatan {
  font-size: 12.5px;
  color: var(--text-2);
  margin-top: 4px;
}
.kriteria-list li {
  display: flex;
  align-items: flex-start;
  gap: 10px;
  font-size: 14px;
  padding: 8px 0;
}
/* Form edit isi */
.form-tambah {
  border: 1px solid var(--border-kuat);
  border-radius: var(--radius-sm);
  background: var(--bg);
  padding: 14px;
  margin-bottom: 14px;
  display: flex;
  flex-direction: column;
  gap: 10px;
}
.form-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
}
.form-baris {
  display: flex;
  gap: 10px;
}
.inp {
  height: 38px;
  padding: 0 10px;
  border: 1px solid var(--border-kuat);
  border-radius: 8px;
  background: var(--surface);
  font-size: 13.5px;
  font-family: inherit;
  color: var(--text);
  min-width: 0;
}
.inp:focus {
  outline: none;
  border-color: var(--brand);
  box-shadow: 0 0 0 3px rgba(242, 105, 94, 0.15);
}
.inp.lebar {
  grid-column: span 2;
  flex: 1;
}
.inp.sempit {
  max-width: 90px;
}
.form-aksi {
  display: flex;
  justify-content: flex-end;
  gap: 8px;
}
.form-aksi.span {
  grid-column: 1 / -1;
}
.pesan-edit {
  display: flex;
  align-items: center;
  gap: 6px;
  color: var(--merah);
  font-size: 12.5px;
  margin-bottom: 12px;
}
.hapus-btn {
  border: none;
  background: transparent;
  color: var(--text-3);
  display: grid;
  place-items: center;
  width: 30px;
  height: 30px;
  border-radius: 8px;
  flex-shrink: 0;
  transition: background 0.14s, color 0.14s;
}
.hapus-btn:hover {
  background: var(--merah-bg);
  color: var(--merah);
}
.acuan-isi {
  flex: 1;
  min-width: 0;
}
.kriteria-uraian {
  flex: 1;
}
.kriteria-list li {
  align-items: center;
}

/* Samping */
.samping {
  display: flex;
  flex-direction: column;
  gap: 16px;
  position: sticky;
  top: 84px;
}
.menunggu {
  font-size: 13px;
  color: var(--text-2);
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 6px;
  margin-bottom: 14px;
}
.peran {
  background: var(--kuning-bg);
  color: var(--kuning);
}
.aktif-note {
  display: flex;
  align-items: center;
  gap: 8px;
  color: var(--hijau);
  font-size: 13px;
  font-weight: 600;
  margin-bottom: 12px;
}
.aksi-wrap {
  display: flex;
  flex-direction: column;
  gap: 8px;
}
.aksi-btn {
  width: 100%;
  height: 42px;
}
.konfirmasi {
  border: 1px solid var(--border-kuat);
  border-radius: var(--radius-sm);
  padding: 12px;
  background: var(--bg);
}
.konf-judul {
  font-weight: 700;
  font-size: 14px;
  margin-bottom: 8px;
}
.konfirmasi textarea {
  width: 100%;
  border: 1px solid var(--border-kuat);
  border-radius: 8px;
  padding: 8px;
  font-family: inherit;
  font-size: 13px;
  background: var(--surface);
  resize: vertical;
  margin-bottom: 8px;
}
.konf-galat {
  display: flex;
  align-items: center;
  gap: 6px;
  color: var(--merah);
  font-size: 12px;
  margin-bottom: 8px;
}
.konf-aksi {
  display: flex;
  justify-content: flex-end;
  gap: 8px;
}
.btn-sm {
  height: 34px;
  padding: 0 14px;
  font-size: 13px;
}
.no-aksi {
  font-size: 13px;
  color: var(--text-3);
}
.syarat-list {
  list-style: none;
  padding: 0;
  display: flex;
  flex-direction: column;
  gap: 12px;
}
.syarat-list li {
  display: flex;
  gap: 10px;
}
.syarat-ikon {
  width: 20px;
  height: 20px;
  flex-shrink: 0;
  border-radius: 50%;
  display: grid;
  place-items: center;
  margin-top: 1px;
}
.syarat-ikon.ya {
  background: var(--hijau-bg);
  color: var(--hijau);
}
.syarat-ikon.tidak {
  background: var(--kuning-bg);
  color: var(--kuning);
}
.syarat-judul {
  font-size: 13px;
  font-weight: 600;
  display: flex;
  align-items: center;
  gap: 6px;
}
.mini-hambat {
  background: var(--merah-bg);
  color: var(--merah);
  font-size: 9.5px;
  padding: 1px 6px;
}
.syarat-pj {
  font-size: 11.5px;
  color: var(--text-3);
  margin-top: 2px;
}
.kekurangan {
  margin: 6px 0 0;
  padding-left: 16px;
  font-size: 12px;
  color: var(--kuning);
}
.kekurangan li {
  display: list-item;
  list-style: disc;
}
/* Timeline */
.timeline {
  list-style: none;
  padding: 0;
  margin: 0;
}
.timeline li {
  position: relative;
  padding: 0 0 18px 24px;
  border-left: 2px solid var(--border);
}
.timeline li:last-child {
  border-left-color: transparent;
  padding-bottom: 0;
}
.tl-dot {
  position: absolute;
  left: -7px;
  top: 2px;
  width: 12px;
  height: 12px;
  border-radius: 50%;
  background: var(--brand);
  border: 2px solid var(--surface);
}
.tl-aksi {
  font-weight: 700;
  font-size: 13.5px;
  display: flex;
  align-items: center;
  gap: 8px;
}
.tl-meta {
  font-size: 12px;
  color: var(--text-3);
  margin-top: 3px;
}
.tl-alasan {
  font-size: 12.5px;
  color: var(--text-2);
  font-style: italic;
  margin-top: 4px;
}
@media (max-width: 920px) {
  .kolom {
    grid-template-columns: 1fr;
  }
  .samping {
    position: static;
  }
}
</style>
