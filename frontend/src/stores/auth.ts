import { defineStore } from 'pinia'
import { api, ambilToken, simpanToken } from '@/lib/api'

export interface Pengguna {
  id: number
  nama: string
  email: string
  peran: string
  peran_label: string
}

const ADMIN = ['ADMIN_RS', 'SYSTEM_ADMIN', 'SUPER_ADMIN']

export const useAuth = defineStore('auth', {
  state: () => ({
    pengguna: null as Pengguna | null,
    token: ambilToken(),
    siap: false, // sudah selesai cek sesi awal
  }),
  getters: {
    masuk: (s) => !!s.token,
    isAdmin: (s) => ADMIN.includes(s.pengguna?.peran ?? ''),
  },
  actions: {
    async login(email: string, password: string) {
      const res = await api<{ token: string; user: Pengguna }>('auth/login', {
        method: 'POST',
        body: JSON.stringify({ email, password }),
      })
      this.token = res.token
      this.pengguna = res.user
      simpanToken(res.token)
    },

    async muatProfil() {
      if (!this.token) {
        this.siap = true
        return
      }
      try {
        const res = await api<{ user: Pengguna }>('auth/me')
        this.pengguna = res.user
      } catch {
        this.bersihkan()
      } finally {
        this.siap = true
      }
    },

    async logout() {
      try {
        await api('auth/logout', { method: 'POST' })
      } catch {
        /* abaikan */
      } finally {
        this.bersihkan()
      }
    },

    bersihkan() {
      this.token = null
      this.pengguna = null
      simpanToken(null)
    },
  },
})
