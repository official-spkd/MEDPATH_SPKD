<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { BookOpenText, Bookmark } from 'lucide-vue-next'
import { api } from '@/lib/api'
import { LABEL_SUMBER } from '@/lib/format'
import PageHeader from '@/components/PageHeader.vue'

interface Dokumen {
  id: number
  nama: string
  sumber: string
  penerbit: string
  tahun: number
  icd10_cakupan: string[]
  butir_count: number
}

const data = ref<Dokumen[]>([])
onMounted(async () => {
  const res = await api<{ data: Dokumen[] }>('dokumen-panduan')
  data.value = res.data
})

const warnaSumber: Record<string, string> = {
  PNPK: 'info',
  PPK_RS: 'brand',
  PPK_ASOSIASI: 'ungu',
}
</script>

<template>
  <div>
    <PageHeader judul="Dokumen Panduan" demo sub="PPK, PNPK, dan panduan asosiasi profesi — sumber acuan klinis yang butirnya diekstrak untuk menyusun CP.">
      <template #aksi><button class="btn btn-brand"><BookOpenText :size="16" /> Tambah Dokumen</button></template>
    </PageHeader>

    <div class="grid">
      <article v-for="dk in data" :key="dk.id" class="card dok">
        <div class="dok-head">
          <span class="dok-ikon"><BookOpenText :size="18" /></span>
          <span class="badge" :class="`s-${warnaSumber[dk.sumber] ?? 'brand'}`">{{ LABEL_SUMBER[dk.sumber] ?? dk.sumber }}</span>
        </div>
        <h3 class="dok-nama">{{ dk.nama }}</h3>
        <p class="dok-penerbit">{{ dk.penerbit }} · {{ dk.tahun }}</p>
        <div class="cakupan">
          <span v-for="c in dk.icd10_cakupan" :key="c" class="mono chip-icd">{{ c }}</span>
        </div>
        <div class="dok-foot">
          <Bookmark :size="14" />
          <span class="angka">{{ dk.butir_count }}</span> butir acuan terekstrak
        </div>
      </article>
    </div>
  </div>
</template>

<style scoped>
.grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(290px, 1fr));
  gap: 16px;
}
.dok {
  padding: 18px 20px;
  display: flex;
  flex-direction: column;
  transition: box-shadow 0.16s, transform 0.16s, border-color 0.16s;
}
.dok:hover {
  box-shadow: var(--shadow-md);
  transform: translateY(-2px);
  border-color: var(--border-kuat);
}
.dok-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 14px;
}
.dok-ikon {
  width: 38px;
  height: 38px;
  border-radius: 10px;
  display: grid;
  place-items: center;
  background: var(--brand-light);
  color: var(--brand-aktif);
}
.s-info {
  background: var(--info-bg);
  color: var(--info);
}
.s-brand {
  background: var(--brand-light);
  color: var(--brand-aktif);
}
.s-ungu {
  background: var(--ungu-bg);
  color: var(--ungu);
}
.dok-nama {
  font-size: 15px;
  font-weight: 700;
  line-height: 1.4;
}
.dok-penerbit {
  font-size: 12.5px;
  color: var(--text-3);
  margin-top: 4px;
}
.cakupan {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
  margin: 14px 0;
}
.chip-icd {
  font-size: 11px;
  font-weight: 600;
  padding: 2px 8px;
  border-radius: 6px;
  background: var(--bg);
  border: 1px solid var(--border);
  color: var(--text-2);
}
.dok-foot {
  margin-top: auto;
  padding-top: 12px;
  border-top: 1px solid var(--border);
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 12.5px;
  color: var(--text-2);
}
.dok-foot .angka {
  font-weight: 700;
  color: var(--text);
}
</style>
