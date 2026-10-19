<script setup lang="ts">
import { ref, watch, computed } from 'vue'
import { useRoute } from 'vue-router'
import { api } from '@/lib/api'
import PageHeader from '@/components/PageHeader.vue'
const route = useRoute()
const cp = ref<{kode:string;nama:string}[]>([])
const dokumen = ref<{id:number;nama:string}[]>([])
const galat = ref('')
const q = computed(() => String(route.query.q ?? '').trim().toLocaleLowerCase('id'))
watch(q, async () => {
  try {
    const [a,b] = await Promise.all([api<{data:typeof cp.value}>('cp'), api<{data:typeof dokumen.value}>('dokumen-panduan')])
    cp.value = a.data.filter(x => `${x.kode} ${x.nama}`.toLowerCase().includes(q.value))
    dokumen.value = b.data.filter(x => x.nama.toLowerCase().includes(q.value))
  } catch { galat.value = 'Pencarian belum dapat dimuat. Coba lagi.' }
}, { immediate: true })
</script>
<template><div><PageHeader judul="Hasil pencarian" :sub="String(route.query.q ?? '')" />
<p v-if="galat" role="alert">{{ galat }}</p>
<section class="card hasil"><h2>Clinical Pathway ({{ cp.length }})</h2><router-link v-for="c in cp" :key="c.kode" :to="`/library/${c.kode}`">{{ c.kode }} — {{ c.nama }}</router-link><p v-if="!cp.length">Tidak ada CP yang sesuai.</p></section>
<section class="card hasil"><h2>Dokumen ({{ dokumen.length }})</h2><router-link v-for="d in dokumen" :key="d.id" :to="`/dokumen/${d.id}`">{{ d.nama }}</router-link><p v-if="!dokumen.length">Tidak ada dokumen yang sesuai.</p></section>
</div></template>
<style scoped>.hasil{padding:20px;margin-bottom:16px}.hasil a{display:block;padding:12px 0;color:var(--brand-aktif)}</style>
