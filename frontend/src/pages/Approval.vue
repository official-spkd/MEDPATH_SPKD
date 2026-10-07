<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { Stamp } from 'lucide-vue-next'
import { api } from '@/lib/api'
import { angka } from '@/lib/format'
import PageHeader from '@/components/PageHeader.vue'
import StatusPill from '@/components/StatusPill.vue'

interface CP {
  kode: string
  nama: string
  kelompok_diagnosis: string
  status: string
  jumlah_episode: number
  menunggu_peran: string
  status_sejak: string
}

const data = ref<CP[]>([])
const memuat = ref(true)

onMounted(async () => {
  try {
    const res = await api<{ data: CP[] }>('approval')
    data.value = res.data
  } finally {
    memuat.value = false
  }
})
</script>

<template>
  <div>
    <PageHeader judul="Approval" demo sub="Antrean pengesahan Clinical Pathway — Tim CP, Komite Medik, dan Direktur. Setiap tahap tercatat dalam audit trail." />

    <div class="card panel">
      <div class="panel-head">
        <div class="jml-wrap">
          <span class="jml-ikon"><Stamp :size="18" /></span>
          <p class="jml angka">{{ angka(data.length) }} CP menunggu tindakan</p>
        </div>
      </div>

      <div v-if="!memuat && !data.length" class="kosong">Tidak ada CP dalam antrean pengesahan. 🎉</div>

      <ul v-else class="list">
        <li v-for="c in data" :key="c.kode" class="item">
          <span class="mono kode">{{ c.kode }}</span>
          <div class="isi">
            <p class="nama">{{ c.nama }}</p>
            <p class="meta">{{ c.kelompok_diagnosis }} · {{ angka(c.jumlah_episode) }} episode · sejak {{ c.status_sejak }}</p>
          </div>
          <div class="menunggu">
            <span class="menunggu-l">Menunggu</span>
            <span class="badge peran">{{ c.menunggu_peran }}</span>
          </div>
          <StatusPill :status="c.status" />
          <router-link :to="`/library/${c.kode}`" class="btn btn-brand btn-sm">Tinjau</router-link>
        </li>
      </ul>
    </div>
  </div>
</template>

<style scoped>
.panel {
  padding: 20px 22px;
}
.panel-head {
  margin-bottom: 16px;
}
.jml-wrap {
  display: flex;
  align-items: center;
  gap: 10px;
}
.jml-ikon {
  width: 36px;
  height: 36px;
  border-radius: 10px;
  display: grid;
  place-items: center;
  background: var(--ungu-bg);
  color: var(--ungu);
}
.jml {
  font-family: var(--font-judul);
  font-weight: 700;
  font-size: 16px;
}
.list {
  list-style: none;
  padding: 0;
  display: flex;
  flex-direction: column;
  gap: 10px;
}
.item {
  display: flex;
  align-items: center;
  gap: 16px;
  padding: 14px 16px;
  border: 1px solid var(--border);
  border-radius: var(--radius-sm);
  transition: border-color 0.14s, box-shadow 0.14s;
}
.item:hover {
  border-color: var(--border-kuat);
  box-shadow: var(--shadow);
}
.kode {
  font-size: 12.5px;
  font-weight: 700;
  color: var(--brand-aktif);
  width: 62px;
  flex-shrink: 0;
}
.isi {
  flex: 1;
  min-width: 0;
}
.nama {
  font-weight: 600;
  font-size: 14px;
}
.meta {
  font-size: 12px;
  color: var(--text-3);
  margin-top: 2px;
}
.menunggu {
  display: flex;
  flex-direction: column;
  align-items: flex-end;
  gap: 3px;
}
.menunggu-l {
  font-size: 10.5px;
  color: var(--text-3);
  text-transform: uppercase;
  letter-spacing: 0.5px;
}
.peran {
  background: var(--kuning-bg);
  color: var(--kuning);
}
.btn-sm {
  height: 34px;
  padding: 0 16px;
  font-size: 13px;
}
.kosong {
  padding: 48px;
  text-align: center;
  color: var(--text-3);
}
@media (max-width: 720px) {
  .item {
    flex-wrap: wrap;
    gap: 10px;
  }
}
</style>
