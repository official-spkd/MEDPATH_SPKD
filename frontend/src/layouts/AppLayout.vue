<script setup lang="ts">
import { ref, onMounted, provide, watch } from 'vue'
import { useRoute } from 'vue-router'
import { useFilter } from '@/stores/filter'
import Sidebar from '@/components/Sidebar.vue'
import Topbar from '@/components/Topbar.vue'
import { api } from '@/lib/api'

const antrean = ref(0)
const menuBuka = ref(false)
const route = useRoute()
const filter = useFilter()
watch(() => route.fullPath, () => { menuBuka.value = false })
watch(() => filter.kunci, muatAntrean)

async function muatAntrean() {
  try {
    const res = await api<{ data: unknown[] }>('approval')
    antrean.value = res.data?.length ?? 0
  } catch {
    /* abaikan */
  }
}
provide('refreshAntrean', muatAntrean)
onMounted(muatAntrean)
</script>

<template>
  <div class="shell">
    <button v-if="menuBuka" class="menu-latar" aria-label="Tutup menu navigasi" @click="menuBuka = false" />
    <Sidebar :antrean="antrean" :class="{ 'mobile-buka': menuBuka }" />
    <div class="utama">
      <Topbar @menu="menuBuka = !menuBuka" />
      <main class="isi">
        <router-view :key="filter.kunci" />
      </main>
    </div>
  </div>
</template>

<style scoped>
.menu-latar { position: fixed; inset: 0; z-index: 39; border: 0; background: #0008; }
@media (min-width: 761px) { .menu-latar { display: none; } }
.shell {
  display: flex;
  min-height: 100dvh;
  background: var(--bg);
}
.utama {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
}
.isi {
  min-width: 0;
  flex: 1;
  padding: 28px 32px 48px;
  max-width: 1500px;
  width: 100%;
  margin: 0 auto;
}
@media (max-width: 760px) {
  .isi {
    padding: 20px 16px 40px;
  }
}
</style>
