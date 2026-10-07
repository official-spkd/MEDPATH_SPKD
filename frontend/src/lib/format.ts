// Util format angka & teks (Indonesia).

export function angka(n: number | null | undefined): string {
  return new Intl.NumberFormat('id-ID').format(n ?? 0)
}

// Rupiah ringkas: 2,3 M / 709,1 jt / 450 rb
export function rupiahRingkas(n: number): string {
  const abs = Math.abs(n)
  if (abs >= 1_000_000_000) return `Rp ${bulat(n / 1_000_000_000)} M`
  if (abs >= 1_000_000) return `Rp ${bulat(n / 1_000_000)} jt`
  if (abs >= 1_000) return `Rp ${bulat(n / 1_000)} rb`
  return `Rp ${angka(n)}`
}

function bulat(n: number): string {
  return n.toLocaleString('id-ID', { maximumFractionDigits: 1 })
}

export function persen(bagian: number, total: number): string {
  if (!total) return '0%'
  return `${((bagian / total) * 100).toLocaleString('id-ID', { maximumFractionDigits: 1 })}%`
}

export function inisial(nama: string): string {
  return nama
    .replace(/\(.*\)/, '')
    .trim()
    .split(/\s+/)
    .slice(0, 2)
    .map((k) => k[0])
    .join('')
    .toUpperCase()
}

export function sapaan(): string {
  const jam = new Date().getHours()
  if (jam < 11) return 'Selamat pagi'
  if (jam < 15) return 'Selamat siang'
  if (jam < 18) return 'Selamat sore'
  return 'Selamat malam'
}

export const LABEL_STATUS: Record<string, string> = {
  DRAFT: 'Draft',
  REVISI: 'Perlu Revisi',
  REVIEW_TIM_CP: 'Review Tim CP',
  REVIEW_KOMITE: 'Review Komite Medik',
  MENUNGGU_DIREKTUR: 'Menunggu Direktur',
  AKTIF: 'Aktif',
}

export const LABEL_KATEGORI_BUTIR: Record<string, string> = {
  KRITERIA_DIAGNOSIS: 'Kriteria diagnosis',
  TATA_LAKSANA: 'Tata laksana',
  DOSIS_OBAT: 'Dosis obat',
  INDIKASI_TERAPI: 'Indikasi terapi',
  KONTRAINDIKASI: 'Kontraindikasi',
  MONITORING: 'Monitoring',
  KOMPLIKASI: 'Komplikasi',
  LAINNYA: 'Lainnya',
}

export const LABEL_SUMBER: Record<string, string> = {
  PPK_RS: 'PPK RS',
  PNPK: 'PNPK',
  PPK_ASOSIASI: 'PPK Asosiasi',
  TIM_CP: 'Tim CP',
  LAINNYA: 'Lainnya',
}
