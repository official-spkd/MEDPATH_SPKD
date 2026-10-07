<script setup lang="ts">
import { computed } from 'vue'
import { LABEL_STATUS } from '@/lib/format'

const props = defineProps<{ status: string }>()

const gaya: Record<string, { bg: string; fg: string }> = {
  AKTIF: { bg: 'var(--hijau-bg)', fg: 'var(--hijau)' },
  DRAFT: { bg: 'var(--navy-100)', fg: 'var(--navy-600)' },
  REVISI: { bg: 'var(--merah-bg)', fg: 'var(--merah)' },
  REVIEW_TIM_CP: { bg: 'var(--info-bg)', fg: 'var(--info)' },
  REVIEW_KOMITE: { bg: 'var(--ungu-bg)', fg: 'var(--ungu)' },
  MENUNGGU_DIREKTUR: { bg: 'var(--kuning-bg)', fg: 'var(--kuning)' },
}

const g = computed(() => gaya[props.status] ?? gaya.DRAFT)
const label = computed(() => LABEL_STATUS[props.status] ?? props.status)
</script>

<template>
  <span class="badge" :style="{ background: g.bg, color: g.fg }">
    <span class="titik" :style="{ background: g.fg }" />
    {{ label }}
  </span>
</template>

<style scoped>
.titik {
  width: 6px;
  height: 6px;
  border-radius: 50%;
}
</style>
