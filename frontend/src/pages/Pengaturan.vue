<script setup lang="ts">
import { computed } from 'vue'
import { Users, Building2, KeyRound, BookMarked, ArrowRight } from 'lucide-vue-next'
import PageHeader from '@/components/PageHeader.vue'
import { useAuth } from '@/stores/auth'

interface Kartu {
  icon: any
  judul: string
  teks: string
  to?: string
  superAdmin?: boolean
}

const auth = useAuth()
const kartu = computed<Kartu[]>(() => {
  const k: Kartu[] = [
    { icon: Users, judul: 'Pengguna & Peran', teks: 'Kelola akun, peran RBAC, dan status aktif pengguna rumah sakit.' },
    { icon: Building2, judul: 'Profil Rumah Sakit', teks: 'Tipe RS, kepemilikan, regional, dan kelas tertinggi untuk resolusi tarif.' },
    { icon: KeyRound, judul: 'Kunci Integrasi', teks: 'Terbitkan API key untuk TransCPR-X mengambil CP aktif.' },
  ]
  if (auth.isSuperAdmin) {
    k.push({
      icon: BookMarked,
      judul: 'Panduan Lengkap Super Admin',
      teks: 'Dokumentasi teknis menyeluruh: arsitektur, frontend, backend, database, API, keamanan, deploy, dan troubleshooting.',
      to: '/pengaturan/panduan',
      superAdmin: true,
    })
  }
  return k
})
</script>

<template>
  <div>
    <PageHeader judul="Pengaturan" sub="Administrasi sistem — hanya untuk Admin RS, System Admin, dan Super Admin." />
    <p class="masuk-sbg muted">Masuk sebagai <strong>{{ auth.pengguna?.peran_label }}</strong></p>
    <div class="grid">
      <component
        :is="k.to ? 'router-link' : 'article'"
        v-for="k in kartu"
        :key="k.judul"
        :to="k.to"
        class="card sel"
        :class="{ khusus: k.superAdmin }"
      >
        <div class="atas">
          <span class="ikon"><component :is="k.icon" :size="20" /></span>
          <span v-if="k.superAdmin" class="badge label-sa">Khusus Super Admin</span>
        </div>
        <h3>{{ k.judul }}</h3>
        <p>{{ k.teks }}</p>
        <span v-if="k.to" class="buka">Buka panduan <ArrowRight :size="14" /></span>
      </component>
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
  display: block;
  padding: 20px;
  cursor: pointer;
  transition: box-shadow 0.16s, transform 0.16s, border-color 0.16s;
}
.sel:hover {
  box-shadow: var(--shadow-md);
  transform: translateY(-2px);
  border-color: var(--border-kuat);
}
.khusus {
  border-color: var(--brand);
  background: linear-gradient(180deg, var(--brand-light), var(--surface) 70%);
}
.atas {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 8px;
  margin-bottom: 12px;
}
.ikon {
  width: 44px;
  height: 44px;
  border-radius: 12px;
  display: grid;
  place-items: center;
  background: var(--brand-light);
  color: var(--brand-aktif);
}
.khusus .ikon {
  background: var(--gradient-brand);
  color: var(--brand);
}
.label-sa {
  background: var(--brand-aktif);
  color: #fff;
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
.buka {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  margin-top: 14px;
  font-size: 13px;
  font-weight: 700;
  color: var(--brand-aktif);
}
</style>
