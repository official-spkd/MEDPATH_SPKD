<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { Search, ArrowRight } from 'lucide-vue-next'
import { api } from '@/lib/api'
import { angka, LABEL_STATUS } from '@/lib/format'
import PageHeader from '@/components/PageHeader.vue'
import StatusPill from '@/components/StatusPill.vue'

interface CP {
  kode: string
  nama: string
  cmg: string
  kelompok_diagnosis: string
  status: string
  jumlah_episode: number
  jumlah_acuan: number
  jumlah_rencana: number
  los_median: number
}

const semua = ref<CP[]>([])
const cari = ref('')
const status = ref('')
const memuat = ref(true)

onMounted(async () => {
  try {
    const res = await api<{ data: CP[] }>('cp')
    semua.value = res.data
  } finally {
    memuat.value = false
  }
})

const hasil = computed(() =>
  semua.value.filter((c) => {
    const q = cari.value.trim().toLowerCase()
    const cocok = !q || c.kode.toLowerCase().includes(q) || c.nama.toLowerCase().includes(q) || c.kelompok_diagnosis.toLowerCase().includes(q)
    const cocokStatus = !status.value || c.status === status.value
    return cocok && cocokStatus
  }),
)
</script>

<template>
  <div>
    <PageHeader judul="CP Library" demo sub="Seluruh Clinical Pathway yang disusun dari pola klaim INA-CBG — telusuri, saring status, dan buka detail." />

    <!-- Filter -->
    <div class="card filter-bar">
      <div class="cari">
        <Search :size="17" class="cari-ikon" />
        <input v-model="cari" type="text" placeholder="Cari kode, nama CP, atau kelompok diagnosis…" />
      </div>
      <select v-model="status" class="filter">
        <option value="">Semua status</option>
        <option v-for="(l, k) in LABEL_STATUS" :key="k" :value="k">{{ l }}</option>
      </select>
    </div>

    <!-- Tabel -->
    <div class="card panel">
      <div class="panel-head">
        <p class="jml angka">{{ angka(hasil.length) }} Clinical Pathway</p>
      </div>
      <div class="tabel-bungkus">
        <table class="tabel">
          <thead>
            <tr>
              <th>Kode</th>
              <th>Clinical Pathway</th>
              <th>Kelompok</th>
              <th class="kanan">Episode</th>
              <th class="kanan">LOS</th>
              <th class="kanan">Acuan</th>
              <th class="kanan">Rencana</th>
              <th>Status</th>
              <th></th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="c in hasil" :key="c.kode">
              <td><span class="mono kode">{{ c.kode }}</span></td>
              <td class="nama">{{ c.nama }}</td>
              <td class="muted">{{ c.kelompok_diagnosis }}</td>
              <td class="kanan angka tebal">{{ angka(c.jumlah_episode) }}</td>
              <td class="kanan angka muted">{{ c.los_median || '—' }}</td>
              <td class="kanan">
                <span v-if="c.jumlah_acuan" class="badge ok angka">{{ c.jumlah_acuan }}</span>
                <span v-else class="badge kurang">belum</span>
              </td>
              <td class="kanan angka muted">{{ c.jumlah_rencana || '—' }}</td>
              <td><StatusPill :status="c.status" /></td>
              <td class="kanan"><router-link :to="`/library/${c.kode}`" class="btn btn-outline btn-sm">Detail <ArrowRight :size="14" /></router-link></td>
            </tr>
            <tr v-if="!memuat && !hasil.length">
              <td colspan="9" class="kosong">Tidak ada CP yang cocok dengan filter.</td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
  </div>
</template>

<style scoped>
.filter-bar {
  display: flex;
  gap: 12px;
  padding: 14px;
  margin-bottom: 16px;
  flex-wrap: wrap;
}
.cari {
  flex: 1;
  min-width: 220px;
  display: flex;
  align-items: center;
  gap: 8px;
  height: 42px;
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
.filter {
  height: 42px;
  padding: 0 12px;
  border: 1px solid var(--border);
  border-radius: var(--radius-sm);
  background: var(--surface);
  color: var(--text-2);
  font-size: 14px;
  font-weight: 600;
}
.panel {
  padding: 18px 22px;
}
.panel-head {
  margin-bottom: 14px;
}
.jml {
  font-family: var(--font-judul);
  font-weight: 700;
  font-size: 16px;
}
.btn-sm {
  height: 30px;
  padding: 0 10px;
  font-size: 12.5px;
}
.tabel-bungkus {
  overflow-x: auto;
  margin: 0 -22px -18px;
}
.tabel {
  width: 100%;
  border-collapse: collapse;
  font-size: 13.5px;
  white-space: nowrap;
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
.tabel th:first-child,
.tabel td:first-child {
  padding-left: 22px;
}
.tabel th:last-child,
.tabel td:last-child {
  padding-right: 22px;
}
.tabel td {
  padding: 12px 14px;
  border-bottom: 1px solid var(--border);
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
.ok {
  background: var(--hijau-bg);
  color: var(--hijau);
}
.kurang {
  background: var(--kuning-bg);
  color: var(--kuning);
}
.kosong {
  text-align: center;
  padding: 40px;
  color: var(--text-3);
}
</style>
