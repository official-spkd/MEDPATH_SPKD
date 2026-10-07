import { defineStore } from 'pinia'

const KUNCI = 'transcpg-tema'

function awal(): 'light' | 'dark' {
  try {
    const t = localStorage.getItem(KUNCI)
    if (t === 'dark' || t === 'light') return t
  } catch {
    /* abaikan */
  }
  return matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light'
}

export const useTema = defineStore('tema', {
  state: () => ({ mode: awal() as 'light' | 'dark' }),
  getters: { gelap: (s) => s.mode === 'dark' },
  actions: {
    terapkan() {
      document.documentElement.classList.toggle('dark', this.mode === 'dark')
    },
    ganti() {
      this.mode = this.mode === 'dark' ? 'light' : 'dark'
      try {
        localStorage.setItem(KUNCI, this.mode)
      } catch {
        /* abaikan */
      }
      this.terapkan()
    },
  },
})
