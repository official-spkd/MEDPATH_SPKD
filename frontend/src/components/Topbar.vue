<script setup lang="ts">
import { computed } from 'vue'
import { useRoute } from 'vue-router'
import { Search, Bell, Moon, Sun, ChevronDown } from 'lucide-vue-next'
import { useTema } from '@/stores/tema'
import { useAuth } from '@/stores/auth'
import { inisial } from '@/lib/format'

const route = useRoute()
const tema = useTema()
const auth = useAuth()

const judul = computed(() => (route.meta.judul as string) ?? 'MedPath')
</script>

<template>
  <header class="topbar">
    <div class="kiri">
      <span class="brand">MedPath</span>
      <span class="sep">/</span>
      <span class="halaman">{{ judul }}</span>
    </div>

    <div class="cari">
      <Search :size="16" class="cari-ikon" />
      <input type="text" placeholder="Cari CP, kode INA-CBG, dokumen…" />
      <kbd>Ctrl K</kbd>
    </div>

    <div class="kanan">
      <select class="filter">
        <option>Semua RS</option>
        <option>RSUD Demo SPKD</option>
      </select>
      <select class="filter sembunyi-kecil">
        <option>Semua Layanan</option>
        <option>Rawat Inap</option>
        <option>Rawat Jalan</option>
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
