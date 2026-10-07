<script setup lang="ts">
defineProps<{ judul: string; sub?: string; demo?: boolean }>()
</script>

<template>
  <div class="ph">
    <div>
      <div class="baris">
        <h1>{{ judul }}</h1>
        <span v-if="demo" class="badge data-demo"><span class="titik" /> Data Demo</span>
      </div>
      <p v-if="sub" class="sub">{{ sub }}</p>
    </div>
    <div class="aksi"><slot name="aksi" /></div>
  </div>
</template>

<style scoped>
.ph {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 16px;
  flex-wrap: wrap;
  margin-bottom: 20px;
}
.baris {
  display: flex;
  align-items: center;
  gap: 12px;
}
h1 {
  font-size: 26px;
  font-weight: 800;
  letter-spacing: -0.5px;
}
.sub {
  margin-top: 6px;
  color: var(--text-2);
  font-size: 14px;
  max-width: 70ch;
}
.data-demo {
  background: var(--kuning-bg);
  color: var(--kuning);
}
.data-demo .titik {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: var(--kuning);
}
</style>
