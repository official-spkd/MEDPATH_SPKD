<script setup lang="ts">
import { onMounted } from 'vue'
import { useTema } from '@/stores/tema'

const tema = useTema()
onMounted(() => tema.terapkan())
</script>

<template>
  <router-view />
</template>
