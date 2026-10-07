<script setup lang="ts">
import { ref, onMounted, provide } from 'vue'
import Sidebar from '@/components/Sidebar.vue'
import Topbar from '@/components/Topbar.vue'
import { api } from '@/lib/api'

const antrean = ref(0)

async function muatAntrean() {
  try {
    const res = await api<{ data: unknown[] }>('approval')
    antrean.value = res.data.length
  } catch {
    /* abaikan */
  }
}
provide('refreshAntrean', muatAntrean)
onMounted(muatAntrean)
</script>

<template>
  <div class="shell">
    <Sidebar :antrean="antrean" />
    <div class="utama">
      <Topbar />
      <main class="isi">
        <router-view />
      </main>
    </div>
  </div>
</template>

<style scoped>
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
