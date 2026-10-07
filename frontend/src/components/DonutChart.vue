<script setup lang="ts">
import { computed } from 'vue'

interface Segmen {
  label: string
  nilai: number
  warna: string
}
const props = defineProps<{ segmen: Segmen[]; pusatAngka: string; pusatLabel: string }>()

const total = computed(() => props.segmen.reduce((a, s) => a + s.nilai, 0) || 1)
const arc = computed(() => {
  let akum = 0
  return props.segmen.map((s) => {
    const pct = (s.nilai / total.value) * 100
    const seg = { ...s, pct, offset: -akum }
    akum += pct
    return seg
  })
})
</script>

<template>
  <div class="donut-wrap">
    <div class="donut">
      <svg viewBox="0 0 120 120">
        <g transform="rotate(-90 60 60)">
          <circle cx="60" cy="60" r="48" fill="none" stroke="var(--border)" stroke-width="16" pathLength="100" />
          <circle
            v-for="(s, i) in arc"
            :key="i"
            cx="60"
            cy="60"
            r="48"
            fill="none"
            :stroke="s.warna"
            stroke-width="16"
            pathLength="100"
            :stroke-dasharray="`${s.pct} ${100 - s.pct}`"
            :stroke-dashoffset="s.offset"
          />
        </g>
      </svg>
      <div class="pusat">
        <p class="angka pusat-nilai">{{ pusatAngka }}</p>
        <p class="pusat-label">{{ pusatLabel }}</p>
      </div>
    </div>
    <ul class="legenda">
      <li v-for="(s, i) in arc" :key="i">
        <span class="swatch" :style="{ background: s.warna }" />
        <span class="l-label">{{ s.label }}</span>
        <span class="l-nilai angka">{{ s.nilai }}</span>
        <span class="l-pct angka">{{ s.pct.toLocaleString('id-ID', { maximumFractionDigits: 1 }) }}%</span>
      </li>
    </ul>
  </div>
</template>

<style scoped>
.donut-wrap {
  display: flex;
  align-items: center;
  gap: 24px;
  flex-wrap: wrap;
}
.donut {
  position: relative;
  width: 170px;
  height: 170px;
  flex-shrink: 0;
}
.donut svg {
  width: 100%;
  height: 100%;
}
.pusat {
  position: absolute;
  inset: 0;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
}
.pusat-nilai {
  font-family: var(--font-judul);
  font-size: 28px;
  font-weight: 800;
  color: var(--text);
}
.pusat-label {
  font-size: 11px;
  color: var(--text-3);
}
.legenda {
  list-style: none;
  padding: 0;
  flex: 1;
  min-width: 180px;
}
.legenda li {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 6px 0;
  font-size: 13px;
}
.swatch {
  width: 11px;
  height: 11px;
  border-radius: 3px;
  flex-shrink: 0;
}
.l-label {
  flex: 1;
  color: var(--text-2);
}
.l-nilai {
  font-weight: 700;
  color: var(--text);
}
.l-pct {
  width: 48px;
  text-align: right;
  color: var(--text-3);
  font-size: 12px;
}
</style>
