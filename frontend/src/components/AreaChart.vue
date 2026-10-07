<script setup lang="ts">
import { computed } from 'vue'

interface Titik {
  periode: string
  episode: number
  cpAktif: number
}
const props = defineProps<{ data: Titik[] }>()

const W = 760
const H = 260
const P = { t: 16, r: 44, b: 28, l: 44 }
const iw = W - P.l - P.r
const ih = H - P.t - P.b
const uid = `a${Math.random().toString(36).slice(2, 8)}`

const maxEp = computed(() => Math.max(1, ...props.data.map((d) => d.episode)) * 1.1)
const maxCp = computed(() => Math.max(1, ...props.data.map((d) => d.cpAktif)) * 1.2)

function x(i: number) {
  return P.l + (props.data.length <= 1 ? iw / 2 : (i / (props.data.length - 1)) * iw)
}
const yEp = (v: number) => P.t + ih - (v / maxEp.value) * ih
const yCp = (v: number) => P.t + ih - (v / maxCp.value) * ih

const garisEp = computed(() => props.data.map((d, i) => `${x(i)},${yEp(d.episode)}`).join(' '))
const areaEp = computed(
  () => `M ${x(0)},${P.t + ih} L ${props.data.map((d, i) => `${x(i)},${yEp(d.episode)}`).join(' L ')} L ${x(props.data.length - 1)},${P.t + ih} Z`,
)
const garisCp = computed(() => props.data.map((d, i) => `${x(i)},${yCp(d.cpAktif)}`).join(' '))
const grid = [0, 0.25, 0.5, 0.75, 1]
</script>

<template>
  <div class="chart">
    <svg :viewBox="`0 0 ${W} ${H}`" preserveAspectRatio="none" class="svg">
      <defs>
        <linearGradient :id="uid" x1="0" y1="0" x2="0" y2="1">
          <stop offset="0%" stop-color="var(--brand)" stop-opacity="0.32" />
          <stop offset="100%" stop-color="var(--brand)" stop-opacity="0.02" />
        </linearGradient>
      </defs>

      <line v-for="g in grid" :key="g" :x1="P.l" :x2="W - P.r" :y1="P.t + ih * g" :y2="P.t + ih * g" class="grid" />

      <path :d="areaEp" :fill="`url(#${uid})`" />
      <polyline :points="garisEp" fill="none" stroke="var(--brand)" stroke-width="2.5" stroke-linejoin="round" stroke-linecap="round" />
      <polyline :points="garisCp" fill="none" stroke="var(--teal)" stroke-width="2.5" stroke-dasharray="6 5" stroke-linejoin="round" stroke-linecap="round" />

      <g v-for="(d, i) in data" :key="i">
        <circle :cx="x(i)" :cy="yEp(d.episode)" r="3.5" fill="var(--brand)" stroke="var(--surface)" stroke-width="1.5" />
        <circle :cx="x(i)" :cy="yCp(d.cpAktif)" r="3" fill="var(--teal)" stroke="var(--surface)" stroke-width="1.5" />
        <text :x="x(i)" :y="H - 8" class="xlabel" text-anchor="middle">{{ d.periode }}</text>
      </g>
    </svg>
  </div>
</template>

<style scoped>
.chart {
  width: 100%;
}
.svg {
  width: 100%;
  height: 260px;
  display: block;
}
.grid {
  stroke: var(--border);
  stroke-width: 1;
  stroke-dasharray: 3 4;
}
.xlabel {
  fill: var(--text-3);
  font-size: 11px;
  font-family: var(--font-teks);
}
</style>
