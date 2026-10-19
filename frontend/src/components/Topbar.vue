<script setup lang="ts">
import { computed, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { Search, Bell, Moon, Sun, ChevronDown, Menu } from 'lucide-vue-next'
import { useFilter } from '@/stores/filter'
import { useTema } from '@/stores/tema'
import { useAuth } from '@/stores/auth'
import { inisial } from '@/lib/format'

const route = useRoute()
const router = useRouter()
const filter = useFilter()
const cari = ref('')
defineEmits<{ menu: [] }>()
function cariGlobal() {
  const q = cari.value.trim()
  if (!q) return
  if (/^[A-Z]-\d-\d{2}$/i.test(q)) router.push(`/library/${q.toUpperCase()}`)
  else router.push({ path: '/pencarian', query: { q } })
}
const tema = useTema()
const auth = useAuth()

const judul = computed(() => (route.meta.judul as string) ?? 'MedPath')
</script>

<template>
  <header class="topbar">
    <button class="ikon-btn menu-mobile" aria-label="Buka menu navigasi" @click="$emit('menu')"><Menu :size="20" /></button>
    <div class="kiri">
      <span class="brand">MedPath</span>
      <span class="sep">/</span>
      <span class="halaman">{{ judul }}</span>
    </div>

    <div class="cari">
      <Search :size="16" class="cari-ikon" />
      <input v-model="cari" aria-label="Pencarian global" type="search" placeholder="Cari CP, kode INA-CBG, dokumen…" @keydown.enter="cariGlobal" />
      <kbd>Ctrl K</kbd>
    </div>

    <div class="kanan">
      <select v-model="filter.rumahSakit" class="filter" aria-label="Rumah sakit">
        <option value="">Semua RS</option>
        <option value="RSUD Demo SPKD">RSUD Demo SPKD</option>
      </select>
      <select v-model="filter.layanan" class="filter sembunyi-kecil" aria-label="Layanan">
        <option value="">Semua Layanan</option>
        <option value="Rawat Inap">Rawat Inap</option>
        <option value="Rawat Jalan">Rawat Jalan</option>
      </select>

      <button class="ikon-btn" :title="tema.gelap ? 'Mode terang' : 'Mode gelap'" @click="tema.ganti()">
        <Sun v-if="tema.gelap" :size="18" />
        <Moon v-else :size="18" />
      </button>

      <button class="ikon-btn" title="Notifikasi">
        <Bell :size="18" />
        <span class="dot" />
      </button>

      <button class="avatar-btn">
        <span class="avatar">{{ inisial(auth.pengguna?.nama ?? '?') }}</span>
        <ChevronDown :size="15" class="muted" />
      </button>
    </div>
  </header>
</template>

<style scoped>
.menu-mobile { display: none !important; }
@media (max-width: 760px) {
  .menu-mobile { display: grid !important; }
  .topbar { height: auto !important; min-height: 64px; padding: 10px !important; gap: 6px !important; flex-wrap: wrap; }
  .kiri { flex: 1; min-width: 0; }
  .brand, .sep, .avatar-btn { display: none !important; }
  .kanan { flex-wrap: wrap; }
  .filter { max-width: 150px; }
  .cari { display: flex !important; order: 3; max-width: none !important; width: 100%; flex-basis: 100% !important; }
  .cari input { min-width: 0; width: 100%; }
}
.topbar {
  height: 64px;
  flex-shrink: 0;
  position: sticky;
  top: 0;
  z-index: 20;
  display: flex;
  align-items: center;
  gap: 16px;
  padding: 0 24px;
  background: color-mix(in srgb, var(--surface) 86%, transparent);
  backdrop-filter: blur(8px);
  border-bottom: 1px solid var(--border);
}
.kiri {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 14px;
}
.brand {
  color: var(--text-3);
  font-weight: 600;
}
.sep {
  color: var(--text-3);
}
.halaman {
  font-weight: 700;
  color: var(--text);
}
.cari {
  flex: 1;
  max-width: 440px;
  margin: 0 auto;
  display: flex;
  align-items: center;
  gap: 8px;
  height: 40px;
  padding: 0 12px;
  background: var(--bg);
  border: 1px solid var(--border);
  border-radius: var(--radius-pill);
}
.cari-ikon {
  color: var(--text-3);
  flex-shrink: 0;
}
.cari input {
  flex: 1;
  border: none;
  outline: none;
  background: transparent;
  font-size: 13.5px;
}
.cari input::placeholder {
  color: var(--text-3);
}
.cari kbd {
  font-size: 10.5px;
  font-weight: 600;
  color: var(--text-3);
  background: var(--surface);
  border: 1px solid var(--border);
  border-radius: 6px;
  padding: 2px 6px;
}
.kanan {
  display: flex;
  align-items: center;
  gap: 8px;
}
.filter {
  height: 38px;
  padding: 0 10px;
  border: 1px solid var(--border);
  border-radius: var(--radius-sm);
  background: var(--surface);
  color: var(--text-2);
  font-size: 13px;
  font-weight: 600;
  cursor: pointer;
}
.ikon-btn {
  position: relative;
  width: 38px;
  height: 38px;
  display: grid;
  place-items: center;
  border: 1px solid var(--border);
  border-radius: var(--radius-sm);
  background: var(--surface);
  color: var(--text-2);
  transition: background 0.14s, color 0.14s;
}
.ikon-btn:hover {
  background: var(--brand-light);
  color: var(--brand-gelap);
}
.dot {
  position: absolute;
  top: 8px;
  right: 9px;
  width: 7px;
  height: 7px;
  border-radius: 50%;
  background: var(--merah);
  border: 1.5px solid var(--surface);
}
.avatar-btn {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 3px 8px 3px 3px;
  border: 1px solid var(--border);
  border-radius: var(--radius-pill);
  background: var(--surface);
}
.avatar {
  width: 32px;
  height: 32px;
  border-radius: 50%;
  background: var(--gradient-brand);
  color: var(--brand);
  font-weight: 800;
  font-size: 12px;
  display: grid;
  place-items: center;
}
@media (max-width: 920px) {
  .sembunyi-kecil {
    display: none;
  }
  .cari {
    display: none;
  }
}
</style>
