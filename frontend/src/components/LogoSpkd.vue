<script setup lang="ts">
import { Activity } from 'lucide-vue-next'
withDefaults(defineProps<{ terang?: boolean; ringkas?: boolean }>(), { terang: false, ringkas: false })
</script>

<template>
  <div class="logo" :class="{ terang }">
    <div class="mark"><Activity :size="22" :stroke-width="2.6" /></div>
    <div v-if="!ringkas" class="teks">
      <p class="nama">MedPath</p>
      <p class="sub">BY SPKD</p>
    </div>
  </div>
</template>

<style scoped>
.logo {
  display: flex;
  align-items: center;
  gap: 12px;
}
.mark {
  width: 42px;
  height: 42px;
  flex-shrink: 0;
  border-radius: 12px;
  display: grid;
  place-items: center;
  color: var(--brand);
  background: var(--gradient-brand);
  box-shadow: 0 6px 16px rgba(150, 102, 15, 0.35), inset 0 1px 0 rgba(255, 255, 255, 0.12);
}
.nama {
  font-family: var(--font-judul);
  font-weight: 800;
  font-size: 19px;
  letter-spacing: -0.3px;
  color: var(--sidebar-teks);
  line-height: 1.1;
}
.sub {
  font-size: 9.5px;
  font-weight: 700;
  letter-spacing: 2px;
  color: var(--sidebar-teks-2);
}
.terang .nama {
  color: #fff;
}
.terang .sub {
  color: rgba(255, 255, 255, 0.6);
}
</style>
