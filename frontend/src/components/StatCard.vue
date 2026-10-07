<script setup lang="ts">
import { computed } from 'vue'

const props = withDefaults(
  defineProps<{
    label: string
    nilai: string
    icon: any
    warna?: 'brand' | 'hijau' | 'merah' | 'kuning' | 'ungu' | 'info'
    keterangan?: string
    to?: string
  }>(),
  { warna: 'brand' },
)

const peta: Record<string, { fg: string; bg: string }> = {
  brand: { fg: 'var(--brand-aktif)', bg: 'var(--brand-light)' },
  hijau: { fg: 'var(--hijau)', bg: 'var(--hijau-bg)' },
  merah: { fg: 'var(--merah)', bg: 'var(--merah-bg)' },
  kuning: { fg: 'var(--kuning)', bg: 'var(--kuning-bg)' },
  ungu: { fg: 'var(--ungu)', bg: 'var(--ungu-bg)' },
  info: { fg: 'var(--info)', bg: 'var(--info-bg)' },
}
const c = computed(() => peta[props.warna])
const tag = computed(() => (props.to ? 'router-link' : 'div'))
</script>

<template>
  <component :is="tag" :to="to" class="stat card" :class="{ tautan: to }">
    <div class="atas">
      <p class="label">{{ label }}</p>
      <span class="ikon" :style="{ color: c.fg, background: c.bg }">
        <component :is="icon" :size="17" :stroke-width="2.2" />
      </span>
    </div>
    <p class="nilai angka" :style="{ color: warna === 'merah' ? 'var(--merah)' : warna === 'hijau' ? 'var(--hijau)' : 'var(--text)' }">
      {{ nilai }}
    </p>
    <p v-if="keterangan" class="ket">{{ keterangan }}</p>
    <div class="garis" :style="{ background: c.fg }" />
  </component>
</template>

<style scoped>
.stat {
  position: relative;
  padding: 16px 16px 18px;
  overflow: hidden;
  display: block;
}
.tautan {
  transition: box-shadow 0.16s, transform 0.16s, border-color 0.16s;
}
.tautan:hover {
  box-shadow: var(--shadow-md);
  transform: translateY(-2px);
  border-color: var(--border-kuat);
}
.atas {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 8px;
}
.label {
  font-size: 11px;
  font-weight: 700;
  letter-spacing: 0.6px;
  text-transform: uppercase;
  color: var(--text-3);
  line-height: 1.4;
}
.ikon {
  width: 32px;
  height: 32px;
  flex-shrink: 0;
  border-radius: 9px;
  display: grid;
  place-items: center;
}
.nilai {
  margin-top: 10px;
  font-family: var(--font-judul);
  font-size: 28px;
  font-weight: 800;
  letter-spacing: -0.5px;
}
.ket {
  margin-top: 4px;
  font-size: 12px;
  color: var(--text-2);
}
.garis {
  position: absolute;
  left: 16px;
  right: 16px;
  bottom: 10px;
  height: 3px;
  border-radius: 999px;
  opacity: 0.55;
}
</style>
