<script setup lang="ts">
import { computed } from 'vue'

const props = withDefaults(defineProps<{ skor: number; maks?: number }>(), { maks: 100 })

// Gauge 270° (celah di bawah). pathLength=100 → dasharray dalam persen.
const BUSUR = 75 // 270° dari 360°
const p = computed(() => Math.max(0, Math.min(1, props.skor / props.maks)) * BUSUR)
const uid = `g${Math.random().toString(36).slice(2, 8)}`
</script>

<template>
  <div class="gauge">
    <svg viewBox="0 0 140 140" class="svg">
      <defs>
        <linearGradient :id="uid" x1="0" y1="0" x2="1" y2="1">
          <stop offset="0%" stop-color="#f2695e" />
          <stop offset="100%" stop-color="#f8a79e" />
        </linearGradient>
      </defs>
      <circle
        cx="70"
        cy="70"
        r="58"
        fill="none"
        class="track"
        stroke-width="12"
        stroke-linecap="round"
        pathLength="100"
        stroke-dasharray="75 25"
        transform="rotate(135 70 70)"
      />
      <circle
        cx="70"
        cy="70"
        r="58"
        fill="none"
        :stroke="`url(#${uid})`"
        stroke-width="12"
        stroke-linecap="round"
        pathLength="100"
        :stroke-dasharray="`${p} ${100 - p}`"
        transform="rotate(135 70 70)"
        class="maju"
      />
    </svg>
    <div class="tengah">
      <p class="nilai angka">{{ Math.round(skor) }}</p>
      <p class="maks">/ {{ maks }}</p>
    </div>
  </div>
</template>

<style scoped>
.gauge {
  position: relative;
  width: 160px;
  height: 160px;
}
.svg {
  width: 100%;
  height: 100%;
}
.track {
  stroke: rgba(255, 255, 255, 0.1);
}
.maju {
  filter: drop-shadow(0 0 6px rgba(242, 105, 94, 0.55));
  transition: stroke-dasharray 0.8s cubic-bezier(0.22, 1, 0.36, 1);
}
.tengah {
  position: absolute;
  inset: 0;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
}
.nilai {
  font-family: var(--font-judul);
  font-size: 42px;
  font-weight: 800;
  color: #fff;
  line-height: 1;
}
.maks {
  font-size: 12px;
  color: rgba(255, 255, 255, 0.55);
  margin-top: 2px;
}
</style>
