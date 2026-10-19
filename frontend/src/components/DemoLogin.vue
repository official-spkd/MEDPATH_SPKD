<script setup lang="ts">
const emit = defineEmits<{ pilih: [email: string, password: string] }>()
const akunDemo = [
  { label: 'Super Admin', email: 'super_admin@demo.local' },
  { label: 'Tim CP', email: 'tim_cp@demo.local' },
  { label: 'DPJP', email: 'dpjp@demo.local' },
  { label: 'Komite Medik', email: 'komite_medik@demo.local' },
  { label: 'Direktur', email: 'direktur@demo.local' },
]
function isiDemo(email: string) { emit('pilih', email, 'Demo-SPKD-2026') }
</script>
<template><div class="demo-panel"><p>Akun demo lokal</p><button v-for="a in akunDemo" :key="a.email" type="button" @click="isiDemo(a.email)">{{a.label}}</button></div></template>
<style scoped>.demo-panel{margin-top:20px;font-size:12px}.demo-panel button{margin:5px;padding:6px 10px;border:1px solid var(--border);border-radius:20px;background:var(--surface);color:var(--text)}</style>
