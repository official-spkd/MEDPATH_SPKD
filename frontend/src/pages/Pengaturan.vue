<script setup lang="ts">
import { Users, Building2, KeyRound } from 'lucide-vue-next'
import PageHeader from '@/components/PageHeader.vue'
import { useAuth } from '@/stores/auth'

const auth = useAuth()
const kartu = [
  { icon: Users, judul: 'Pengguna & Peran', teks: 'Kelola akun, peran RBAC, dan status aktif pengguna rumah sakit.' },
  { icon: Building2, judul: 'Profil Rumah Sakit', teks: 'Tipe RS, kepemilikan, regional, dan kelas tertinggi untuk resolusi tarif.' },
  { icon: KeyRound, judul: 'Kunci Integrasi', teks: 'Terbitkan API key untuk TransCPR-X mengambil CP aktif.' },
]
</script>

<template>
  <div>
    <PageHeader judul="Pengaturan" sub="Administrasi sistem — hanya untuk Admin RS, System Admin, dan Super Admin." />
    <p class="masuk-sbg muted">Masuk sebagai <strong>{{ auth.pengguna?.peran_label }}</strong></p>
    <div class="grid">
      <article v-for="k in kartu" :key="k.judul" class="card sel">
        <span class="ikon"><component :is="k.icon" :size="20" /></span>
        <h3>{{ k.judul }}</h3>
        <p>{{ k.teks }}</p>
      </article>
    </div>
  </div>
</template>

<style scoped>
.masuk-sbg {
  font-size: 13px;
  margin-bottom: 16px;
}
.grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(260px, 1fr));
  gap: 16px;
}
.sel {
  padding: 20px;
  cursor: pointer;
  transition: box-shadow 0.16s, transform 0.16s, border-color 0.16s;
}
.sel:hover {
  box-shadow: var(--shadow-md);
  transform: translateY(-2px);
  border-color: var(--border-kuat);
}
.ikon {
  width: 44px;
  height: 44px;
  border-radius: 12px;
  display: grid;
  place-items: center;
  background: var(--brand-light);
  color: var(--brand-aktif);
  margin-bottom: 12px;
}
.sel h3 {
  font-size: 15px;
}
.sel p {
  margin-top: 6px;
  font-size: 13px;
  color: var(--text-2);
  line-height: 1.5;
}
</style>
