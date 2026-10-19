import { defineStore } from 'pinia'

export const useFilter = defineStore('filter', {
  state: () => ({ rumahSakit: '', layanan: '' }),
  getters: { kunci: (s) => `${s.rumahSakit}:${s.layanan}` },
})
