<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { BadgeCheck, Calendar, FileText } from 'lucide-vue-next'
import { api } from '@/lib/api'
import { angka } from '@/lib/format'
import PageHeader from '@/components/PageHeader.vue'

interface CP {
  kode: string
  nama: string
  kelompok_diagnosis: string
  status: string
  jumlah_episode: number
  jumlah_rencana: number
  los_median: number
  status_sejak: string
}

const semua = ref<CP[]>([])
onMounted(async () => {
  const res = await api<{ data: CP[] }>('cp')
  semua.value = res.data
})
const aktif = computed(() => semua.value.filter((c) => c.status === 'AKTIF'))
</script>

<template>
  <div>
    <PageHeader judul="CP Aktif" demo sub="Clinical Pathway yang telah disahkan Direktur dan dipakai sistem pelayanan (TransCPR-X) lewat API integrasi." />

    <div class="grid">
      <article v-for="c in aktif" :key="c.kode" class="card cp">
        <div class="cp-head">
          <span class="mono kode">{{ c.kode }}</span>
          <span class="badge aktif"><BadgeCheck :size="13" /> Aktif</span>
        </div>
        <h3 class="cp-nama">{{ c.nama }}</h3>
        <p class="cp-kelompok">{{ c.kelompok_diagnosis }}</p>
        <div class="cp-stats">
          <div><p class="angka">{{ angka(c.jumlah_episode) }}</p><span>Episode</span></div>
          <div><p class="angka">{{ c.los_median }}</p><span>LOS median</span></div>
          <div><p class="angka">{{ c.jumlah_rencana }}</p><span>Rencana</span></div>
        </div>
        <div class="cp-foot">
          <span><Calendar :size="13" /> Sejak {{ c.status_sejak }}</span>
          <span><FileText :size="13" /> v1</span>
        </div>
      </article>
    </div>
  </div>
</template>

<style scoped>
.grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
  gap: 16px;
}
.cp {
  padding: 18px 20px;
  transition: box-shadow 0.16s, transform 0.16s, border-color 0.16s;
}
.cp:hover {
  box-shadow: var(--shadow-md);
  transform: translateY(-2px);
  border-color: var(--border-kuat);
}
.cp-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 10px;
}
.kode {
  font-size: 12.5px;
  font-weight: 700;
  color: var(--brand-aktif);
}
.aktif {
  background: var(--hijau-bg);
  color: var(--hijau);
}
.cp-nama {
  font-size: 15px;
  font-weight: 700;
  line-height: 1.35;
}
.cp-kelompok {
  font-size: 12.5px;
  color: var(--text-3);
  margin-top: 3px;
}
.cp-stats {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 8px;
  margin: 16px 0;
  padding: 12px 0;
  border-top: 1px solid var(--border);
  border-bottom: 1px solid var(--border);
  text-align: center;
}
.cp-stats p {
  font-family: var(--font-judul);
  font-weight: 800;
  font-size: 18px;
}
.cp-stats span {
  font-size: 11px;
  color: var(--text-3);
}
.cp-foot {
  display: flex;
  justify-content: space-between;
  font-size: 12px;
  color: var(--text-2);
}
.cp-foot span {
  display: inline-flex;
  align-items: center;
  gap: 5px;
}
</style>
