<script setup lang="ts">
import { computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import {
  LayoutDashboard,
  Library,
  Stamp,
  BadgeCheck,
  TrendingUp,
  BookOpenText,
  ReceiptText,
  Languages,
  Settings,
  LogOut,
  ChevronsUpDown,
} from 'lucide-vue-next'
import { useAuth } from '@/stores/auth'
import { inisial } from '@/lib/format'
import LogoSpkd from './LogoSpkd.vue'

const props = defineProps<{ antrean: number }>()
const auth = useAuth()
const router = useRouter()
const route = useRoute()

// Rute turunan (mis. /library/:kode, /pengaturan/panduan) adalah saudara di router,
// jadi router-link tidak menandainya aktif — cocokkan prefiks path secara manual.
function aktifPrefiks(it: { to: string; exact?: boolean }) {
  return !it.exact && route.path.startsWith(it.to + '/')
}

interface Item {
  label: string
  to: string
  icon: any
  exact?: boolean
  badge?: number
}

const grup = computed<{ judul: string; item: Item[] }[]>(() => {
  const g = [
    {
      judul: 'Utama',
      item: [
        { label: 'Dashboard', to: '/', icon: LayoutDashboard, exact: true },
        { label: 'CP Library', to: '/library', icon: Library },
        { label: 'Approval', to: '/approval', icon: Stamp, badge: props.antrean },
        { label: 'CP Aktif', to: '/cp-aktif', icon: BadgeCheck },
        { label: 'Evaluasi CP', to: '/evaluasi', icon: TrendingUp },
      ] as Item[],
    },
    {
      judul: 'Referensi',
      item: [
        { label: 'Dokumen Panduan', to: '/dokumen', icon: BookOpenText },
        { label: 'Padanan KPTL', to: '/padanan/kptl', icon: ReceiptText },
        { label: 'Padanan SNOMED-CT', to: '/padanan/snomed', icon: Languages },
      ] as Item[],
    },
  ]
  if (auth.isAdmin) {
    g[1].item.push({ label: 'Pengaturan', to: '/pengaturan', icon: Settings })
  }
  return g
})

async function keluar() {
  await auth.logout()
  router.push('/login')
}
</script>

<template>
  <aside class="sidebar">
    <div class="head">
      <LogoSpkd />
    </div>

    <nav class="nav">
      <div v-for="g in grup" :key="g.judul" class="grup">
        <p class="grup-judul">{{ g.judul }}</p>
        <router-link
          v-for="it in g.item"
          :key="it.to"
          :to="it.to"
          class="item"
          :class="{ aktif: aktifPrefiks(it) }"
          :active-class="it.exact ? '' : 'aktif'"
          :exact-active-class="it.exact ? 'aktif' : ''"
        >
          <component :is="it.icon" :size="19" :stroke-width="2.1" class="ikon" />
          <span class="label">{{ it.label }}</span>
          <span v-if="it.badge" class="badge-nav angka">{{ it.badge > 99 ? '99+' : it.badge }}</span>
        </router-link>
      </div>
    </nav>

    <div class="foot">
      <div class="akun">
        <div class="avatar">{{ inisial(auth.pengguna?.nama ?? '?') }}</div>
        <div class="akun-teks">
          <p class="akun-nama">{{ auth.pengguna?.nama?.replace(/\s*\(.*\)/, '') }}</p>
          <p class="akun-peran">{{ auth.pengguna?.peran_label }}</p>
        </div>
        <ChevronsUpDown :size="15" class="akun-chev" />
      </div>
      <button class="keluar" @click="keluar">
        <LogOut :size="17" :stroke-width="2.1" />
        <span>Keluar</span>
      </button>
    </div>
  </aside>
</template>

<style scoped>
.sidebar {
  width: 248px;
  flex-shrink: 0;
  height: 100dvh;
  position: sticky;
  top: 0;
  display: flex;
  flex-direction: column;
  background: var(--sidebar-grad);
  color: var(--sidebar-teks);
}
.head {
  padding: 20px 20px 14px;
}
.nav {
  flex: 1;
  overflow-y: auto;
  padding: 6px 12px 12px;
}
.grup + .grup {
  margin-top: 18px;
}
.grup-judul {
  font-size: 10.5px;
  font-weight: 800;
  letter-spacing: 1.4px;
  text-transform: uppercase;
  color: var(--sidebar-teks-2);
  padding: 8px 12px 6px;
}
.item {
  display: flex;
  align-items: center;
  gap: 11px;
  padding: 9px 12px;
  border-radius: 10px;
  font-weight: 600;
  font-size: 14px;
  color: var(--sidebar-teks);
  margin-bottom: 2px;
  transition: background 0.14s, color 0.14s;
}
.item:hover {
  background: var(--sidebar-hover);
}
.item.aktif {
  background: var(--sidebar-aktif-bg);
  color: var(--sidebar-aktif-teks);
  box-shadow: 0 2px 8px rgba(120, 30, 20, 0.18);
}
.ikon {
  flex-shrink: 0;
  opacity: 0.92;
}
.label {
  flex: 1;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}
.badge-nav {
  background: var(--merah);
  color: #fff;
  font-size: 10.5px;
  font-weight: 800;
  min-width: 20px;
  height: 20px;
  padding: 0 6px;
  border-radius: 999px;
  display: grid;
  place-items: center;
}
.foot {
  padding: 12px;
  border-top: 1px solid var(--sidebar-garis);
}
.akun {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 8px 10px;
  border-radius: 12px;
  background: var(--sidebar-aktif-bg);
  box-shadow: 0 2px 8px rgba(120, 30, 20, 0.14);
}
.avatar {
  width: 34px;
  height: 34px;
  flex-shrink: 0;
  border-radius: 50%;
  background: var(--gradient-brand);
  color: var(--brand);
  font-weight: 800;
  font-size: 12.5px;
  display: grid;
  place-items: center;
}
.akun-teks {
  min-width: 0;
  flex: 1;
}
.akun-nama {
  font-size: 13px;
  font-weight: 700;
  color: var(--navy);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}
.akun-peran {
  font-size: 11px;
  color: var(--brand-aktif);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}
.akun-chev {
  color: var(--sidebar-teks-2);
  flex-shrink: 0;
}
.keluar {
  margin-top: 8px;
  width: 100%;
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 9px 12px;
  border: none;
  background: transparent;
  color: var(--sidebar-teks);
  font-weight: 600;
  font-size: 14px;
  border-radius: 10px;
  transition: background 0.14s;
}
.keluar:hover {
  background: var(--sidebar-hover);
}
</style>
