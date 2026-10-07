<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import {
  Library,
  BookDashed,
  Stamp,
  BadgeCheck,
  BookOpenText,
  Bookmark,
  ArrowRight,
  FilePen,
  TriangleAlert,
} from 'lucide-vue-next'
import { api, GalatApi } from '@/lib/api'
import { angka, persen, sapaan, LABEL_STATUS, LABEL_KATEGORI_BUTIR } from '@/lib/format'
import { useAuth } from '@/stores/auth'
import StatCard from '@/components/StatCard.vue'
import Gauge from '@/components/Gauge.vue'
import DonutChart from '@/components/DonutChart.vue'
import AreaChart from '@/components/AreaChart.vue'
import StatusPill from '@/components/StatusPill.vue'

interface Dashboard {
  kartu: {
    jumlah_cp: number
    cp_tanpa_acuan: number
    dokumen_panduan: number
    butir_acuan: number
    cp_aktif: number
    dalam_pengesahan: number
    total_episode: number
  }
  readiness: { skor: number; label: string; komponen: { nama: string; nilai: number; maks: number }[] }
  kemajuan_acuan: { sudah: number; menunggu: number; dilewati: number }
  per_status: Record<string, number>
  per_kelompok: { kode: string; nama: string; episode: number; jumlah_cp: number }[]
  prioritas: { kode: string; nama: string; kelompok_diagnosis: string; jumlah_episode: number; status: string; jumlah_acuan: number }[]
  butir_per_kategori: Record<string, number>
  tren: { periode: string; episode: number; cp_aktif: number }[]
  tindakan: { ikon: string; judul: string; detail: string; prioritas: string }[]
}

const auth = useAuth()
const d = ref<Dashboard | null>(null)
const galat = ref('')

onMounted(async () => {
  try {
    d.value = await api<Dashboard>('dashboard')
  } catch (e) {
    galat.value = e instanceof GalatApi ? e.message : 'Gagal memuat dashboard.'
  }
})

const URUT_STATUS = ['AKTIF', 'MENUNGGU_DIREKTUR', 'REVIEW_KOMITE', 'REVIEW_TIM_CP', 'REVISI', 'DRAFT']
const WARNA_STATUS: Record<string, string> = {
  AKTIF: 'var(--hijau)',
  MENUNGGU_DIREKTUR: 'var(--kuning)',
  REVIEW_KOMITE: 'var(--ungu)',
  REVIEW_TIM_CP: 'var(--info)',
  REVISI: 'var(--merah)',
  DRAFT: 'var(--text-3)',
}
const IKON_TINDAKAN: Record<string, any> = { stamp: Stamp, 'book-dashed': BookDashed, 'file-pen': FilePen }

const segmenStatus = computed(() =>
  URUT_STATUS.filter((s) => (d.value?.per_status[s] ?? 0) > 0).map((s) => ({
    label: LABEL_STATUS[s],
    nilai: d.value!.per_status[s],
    warna: WARNA_STATUS[s],
  })),
)
const trenData = computed(() => d.value?.tren.map((t) => ({ periode: t.periode, episode: t.episode, cpAktif: t.cp_aktif })) ?? [])
const totalEpisodeKelompok = computed(() => d.value?.per_kelompok.reduce((a, k) => a + k.episode, 0) ?? 0)
const maxButir = computed(() => Math.max(1, ...Object.values(d.value?.butir_per_kategori ?? {})))
</script>

<template>
  <div class="dash">
    <!-- Header -->
    <div class="head">
      <div>
        <div class="judul-baris">
          <h1>Dashboard Eksekutif</h1>
          <span class="badge data-demo"><span class="titik" /> Data Demo</span>
        </div>
        <p class="sapaan">{{ sapaan() }}, {{ auth.pengguna?.nama?.replace(/\s*\(.*\)/, '') }} — kemajuan penyusunan Clinical Pathway dalam 30 detik.</p>
      </div>
      <router-link to="/library" class="btn btn-outline"><Library :size="16" /> Buka CP Library</router-link>
    </div>

    <div v-if="galat" class="galat-box"><TriangleAlert :size="18" /> {{ galat }}</div>

    <div v-if="!d && !galat" class="memuat">Memuat…</div>

    <template v-if="d">
      <!-- Hero readiness -->
      <section class="hero">
        <div class="hero-gauge">
          <Gauge :skor="d.readiness.skor" />
        </div>
        <div class="hero-mid">
          <p class="hero-kap">MedPath Readiness Index</p>
          <p class="hero-label">{{ d.readiness.label }}</p>
          <div class="hero-stats">
            <div class="hs"><span>Total CP</span><strong class="angka">{{ angka(d.kartu.jumlah_cp) }}</strong></div>
            <div class="hs"><span>CP Aktif</span><strong class="angka">{{ angka(d.kartu.cp_aktif) }}</strong></div>
            <div class="hs"><span>Dalam pengesahan</span><strong class="angka">{{ angka(d.kartu.dalam_pengesahan) }}</strong></div>
            <div class="hs"><span>Total episode</span><strong class="angka">{{ angka(d.kartu.total_episode) }}</strong></div>
          </div>
        </div>
        <div class="hero-komponen">
          <p class="hero-kap">Komponen kesiapan</p>
          <div v-for="k in d.readiness.komponen" :key="k.nama" class="komp">
            <div class="komp-baris">
              <span>{{ k.nama }}</span>
              <span class="angka komp-n">{{ k.nilai }}/{{ k.maks }}</span>
            </div>
            <div class="komp-track">
              <div class="komp-maju" :style="{ width: `${Math.min(100, (k.nilai / k.maks) * 100)}%` }" />
            </div>
          </div>
        </div>
        <div class="hero-glow" />
      </section>

      <!-- KPI cards -->
      <section class="kpi">
        <StatCard label="Clinical Pathway" :nilai="angka(d.kartu.jumlah_cp)" :icon="Library" to="/library" keterangan="grouper ≥5 episode" />
        <StatCard label="CP tanpa acuan" :nilai="angka(d.kartu.cp_tanpa_acuan)" :icon="BookDashed" warna="kuning" to="/library" keterangan="perlu ditetapkan Tim CP" />
        <StatCard label="Dalam pengesahan" :nilai="angka(d.kartu.dalam_pengesahan)" :icon="Stamp" warna="ungu" to="/approval" keterangan="Tim CP · Komite · Direktur" />
        <StatCard label="CP aktif" :nilai="angka(d.kartu.cp_aktif)" :icon="BadgeCheck" warna="hijau" to="/cp-aktif" keterangan="memandu pelayanan" />
        <StatCard label="Dokumen panduan" :nilai="angka(d.kartu.dokumen_panduan)" :icon="BookOpenText" warna="info" to="/dokumen" keterangan="PPK · PNPK · asosiasi" />
        <StatCard label="Butir acuan klinis" :nilai="angka(d.kartu.butir_acuan)" :icon="Bookmark" keterangan="terekstrak dari dokumen" />
      </section>

      <!-- Tren + Tindakan -->
      <section class="dua-kolom tren-kolom">
        <div class="card panel">
          <div class="panel-head">
            <div>
              <p class="panel-judul">Tren Episode & CP Aktif</p>
              <p class="panel-sub">Volume episode rawat inap dan jumlah CP aktif per bulan</p>
            </div>
            <div class="legend">
              <span><i class="dot-brand" /> Episode</span>
              <span><i class="dash-teal" /> CP Aktif</span>
            </div>
          </div>
          <AreaChart :data="trenData" />
        </div>

        <div class="card panel">
          <div class="panel-head">
            <p class="panel-judul">Tindakan Prioritas</p>
            <span class="badge count">{{ d.tindakan.length }}</span>
          </div>
          <ul class="tindakan">
            <li v-for="(t, i) in d.tindakan" :key="i">
              <span class="t-ikon" :class="t.prioritas === 'Tinggi' ? 'tinggi' : 'sedang'">
                <component :is="IKON_TINDAKAN[t.ikon] ?? Stamp" :size="17" />
              </span>
              <div class="t-isi">
                <p class="t-judul">{{ t.judul }}</p>
                <p class="t-detail">{{ t.detail }}</p>
              </div>
              <span class="badge" :class="t.prioritas === 'Tinggi' ? 'pri-tinggi' : 'pri-sedang'">{{ t.prioritas }}</span>
            </li>
          </ul>
        </div>
      </section>

      <!-- Donut status + kelompok -->
      <section class="dua-kolom">
        <div class="card panel">
          <p class="panel-judul">Distribusi Status Pengesahan</p>
          <p class="panel-sub">Posisi seluruh CP dalam alur pengesahan</p>
          <div class="donut-pad">
            <DonutChart :segmen="segmenStatus" :pusat-angka="angka(d.kartu.jumlah_cp)" pusat-label="CP" />
          </div>
        </div>

        <div class="card panel">
          <div class="panel-head">
            <div>
              <p class="panel-judul">Kasus per Kelompok Diagnosis</p>
              <p class="panel-sub">Proporsi episode per CMG INA-CBG</p>
            </div>
            <span class="badge count">{{ angka(totalEpisodeKelompok) }} episode</span>
          </div>
          <ul class="bars">
            <li v-for="k in d.per_kelompok.slice(0, 7)" :key="k.kode">
              <div class="bar-baris">
                <span class="mono bar-kode">{{ k.kode }}</span>
                <span class="bar-nama">{{ k.nama }}</span>
                <span class="angka bar-nilai">{{ angka(k.episode) }}</span>
              </div>
              <div class="bar-track">
                <div class="bar-maju" :style="{ width: `${(k.episode / Math.max(totalEpisodeKelompok, 1)) * 100}%` }" />
              </div>
            </li>
          </ul>
        </div>
      </section>

      <!-- Tabel prioritas -->
      <section class="card panel">
        <div class="panel-head">
          <p class="panel-judul">Prioritas Penyusunan CP</p>
          <router-link to="/library" class="btn btn-ghost btn-sm">Semua <ArrowRight :size="15" /></router-link>
        </div>
        <div class="tabel-bungkus">
          <table class="tabel">
            <thead>
              <tr>
                <th>Kode</th>
                <th>Clinical Pathway</th>
                <th>Kelompok diagnosis</th>
                <th class="kanan">Episode</th>
                <th>Status</th>
                <th class="kanan">Acuan</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="p in d.prioritas" :key="p.kode">
                <td><router-link :to="`/library/${p.kode}`" class="mono kode">{{ p.kode }}</router-link></td>
                <td class="nama"><router-link :to="`/library/${p.kode}`" class="nama-link">{{ p.nama }}</router-link></td>
                <td class="muted">{{ p.kelompok_diagnosis }}</td>
                <td class="kanan angka tebal">{{ angka(p.jumlah_episode) }}</td>
                <td><StatusPill :status="p.status" /></td>
                <td class="kanan">
                  <span v-if="p.jumlah_acuan" class="badge ok angka">{{ p.jumlah_acuan }}</span>
                  <span v-else class="badge kurang">belum</span>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </section>
    </template>
  </div>
</template>

<style scoped>
.dash {
  display: flex;
  flex-direction: column;
  gap: 20px;
}
.head {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 16px;
  flex-wrap: wrap;
}
.judul-baris {
  display: flex;
  align-items: center;
  gap: 12px;
}
.head h1 {
  font-size: 26px;
  font-weight: 800;
  letter-spacing: -0.5px;
}
.sapaan {
  margin-top: 6px;
  color: var(--text-2);
  font-size: 14px;
}
.data-demo {
  background: var(--kuning-bg);
  color: var(--kuning);
}
.data-demo .titik {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: var(--kuning);
}
.btn-sm {
  height: 32px;
  padding: 0 10px;
  font-size: 13px;
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
.memuat {
  padding: 60px;
  text-align: center;
  color: var(--text-3);
}

/* Hero */
.hero {
  position: relative;
  overflow: hidden;
  display: grid;
  grid-template-columns: auto 1fr 1.3fr;
  gap: 32px;
  align-items: center;
  padding: 26px 30px;
  border-radius: var(--radius);
  background: var(--gradient-hero);
  color: #fff;
  box-shadow: var(--shadow-md);
}
.hero-gauge {
  display: grid;
  place-items: center;
}
.hero-kap {
  font-size: 11px;
  font-weight: 700;
  letter-spacing: 1px;
  text-transform: uppercase;
  color: var(--brand);
}
.hero-label {
  font-family: var(--font-judul);
  font-size: 19px;
  font-weight: 700;
  margin-top: 4px;
  color: #fff;
}
.hero-stats {
  margin-top: 16px;
  display: flex;
  flex-direction: column;
  gap: 8px;
}
.hs {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  padding: 8px 12px;
  border-radius: 9px;
  background: rgba(255, 255, 255, 0.05);
  border: 1px solid rgba(255, 255, 255, 0.07);
  font-size: 13px;
  color: rgba(255, 255, 255, 0.75);
}
.hs strong {
  color: #fff;
  font-size: 14px;
}
.hero-komponen {
  display: flex;
  flex-direction: column;
  gap: 11px;
}
.komp-baris {
  display: flex;
  justify-content: space-between;
  font-size: 12.5px;
  color: rgba(255, 255, 255, 0.78);
  margin-bottom: 5px;
}
.komp-n {
  color: #fff;
  font-weight: 700;
}
.komp-track {
  height: 6px;
  border-radius: 999px;
  background: rgba(255, 255, 255, 0.1);
  overflow: hidden;
}
.komp-maju {
  height: 100%;
  border-radius: 999px;
  background: linear-gradient(90deg, #d13b30, #f8a79e);
  transition: width 0.8s cubic-bezier(0.22, 1, 0.36, 1);
}
.hero-glow {
  position: absolute;
  right: -80px;
  bottom: -160px;
  width: 360px;
  height: 360px;
  border-radius: 50%;
  border: 1px solid rgba(242, 105, 94, 0.1);
  pointer-events: none;
}

/* KPI */
.kpi {
  display: grid;
  grid-template-columns: repeat(6, 1fr);
  gap: 14px;
}

/* Panels */
.panel {
  padding: 20px 22px;
}
.panel-head {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 12px;
  margin-bottom: 16px;
}
.panel-judul {
  font-family: var(--font-judul);
  font-weight: 700;
  font-size: 16px;
}
.panel-sub {
  font-size: 12.5px;
  color: var(--text-3);
  margin-top: 2px;
}
.dua-kolom {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 16px;
}
.tren-kolom {
  grid-template-columns: 1.7fr 1fr;
}
.legend {
  display: flex;
  gap: 14px;
  font-size: 12px;
  color: var(--text-2);
}
.legend span {
  display: inline-flex;
  align-items: center;
  gap: 6px;
}
.dot-brand {
  width: 10px;
  height: 10px;
  border-radius: 50%;
  background: var(--brand);
}
.dash-teal {
  width: 14px;
  height: 0;
  border-top: 2.5px dashed var(--teal);
}
.count {
  background: var(--brand-light);
  color: var(--brand-aktif);
}

/* Tindakan */
.tindakan {
  list-style: none;
  padding: 0;
  display: flex;
  flex-direction: column;
  gap: 10px;
}
.tindakan li {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px;
  border: 1px solid var(--border);
  border-radius: var(--radius-sm);
  transition: border-color 0.14s, background 0.14s;
}
.tindakan li:hover {
  border-color: var(--border-kuat);
  background: var(--bg);
}
.t-ikon {
  width: 34px;
  height: 34px;
  flex-shrink: 0;
  border-radius: 9px;
  display: grid;
  place-items: center;
}
.t-ikon.tinggi {
  background: var(--merah-bg);
  color: var(--merah);
}
.t-ikon.sedang {
  background: var(--kuning-bg);
  color: var(--kuning);
}
.t-isi {
  flex: 1;
  min-width: 0;
}
.t-judul {
  font-size: 13.5px;
  font-weight: 600;
}
.t-detail {
  font-size: 12px;
  color: var(--text-3);
  margin-top: 2px;
}
.pri-tinggi {
  background: var(--merah-bg);
  color: var(--merah);
}
.pri-sedang {
  background: var(--kuning-bg);
  color: var(--kuning);
}
.donut-pad {
  padding: 8px 0;
}

/* Bars */
.bars {
  list-style: none;
  padding: 0;
  display: flex;
  flex-direction: column;
  gap: 13px;
}
.bar-baris {
  display: flex;
  align-items: baseline;
  gap: 8px;
  font-size: 13px;
}
.bar-kode {
  font-size: 11px;
  font-weight: 700;
  color: var(--brand-aktif);
  width: 18px;
}
.bar-nama {
  flex: 1;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.bar-nilai {
  font-weight: 700;
}
.bar-track {
  margin-top: 5px;
  height: 7px;
  border-radius: 999px;
  background: var(--bg);
  overflow: hidden;
}
.bar-maju {
  height: 100%;
  border-radius: 999px;
  background: linear-gradient(90deg, var(--brand), var(--brand-aktif));
}

/* Tabel */
.tabel-bungkus {
  overflow-x: auto;
  margin: 0 -22px -20px;
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
  letter-spacing: 0.6px;
  text-transform: uppercase;
  color: var(--text-3);
  padding: 10px 14px;
  border-bottom: 1px solid var(--border);
}
.tabel th:first-child {
  padding-left: 22px;
}
.tabel th:last-child {
  padding-right: 22px;
}
.tabel td {
  padding: 12px 14px;
  border-bottom: 1px solid var(--border);
}
.tabel td:first-child {
  padding-left: 22px;
}
.tabel td:last-child {
  padding-right: 22px;
}
.tabel tbody tr:last-child td {
  border-bottom: none;
}
.tabel tbody tr:hover {
  background: var(--bg);
}
.kanan {
  text-align: right;
}
.tebal {
  font-weight: 700;
}
.kode {
  font-size: 12px;
  font-weight: 700;
  color: var(--brand-aktif);
}
.nama {
  font-weight: 600;
}
.nama-link:hover {
  color: var(--brand-aktif);
}
.ok {
  background: var(--hijau-bg);
  color: var(--hijau);
}
.kurang {
  background: var(--kuning-bg);
  color: var(--kuning);
}

@media (max-width: 1180px) {
  .kpi {
    grid-template-columns: repeat(3, 1fr);
  }
  .hero {
    grid-template-columns: 1fr;
    gap: 22px;
  }
  .hero-gauge {
    justify-self: start;
  }
}
@media (max-width: 860px) {
  .dua-kolom,
  .tren-kolom {
    grid-template-columns: 1fr;
  }
  .kpi {
    grid-template-columns: repeat(2, 1fr);
  }
}
</style>
