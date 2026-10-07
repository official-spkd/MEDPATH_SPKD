// Klien API tipis ke backend Go. Token disimpan di localStorage.
const BASE = '/api/v1'
const KUNCI_TOKEN = 'transcpg-token'

export function ambilToken(): string | null {
  try {
    return localStorage.getItem(KUNCI_TOKEN)
  } catch {
    return null
  }
}

export function simpanToken(t: string | null) {
  try {
    if (t) localStorage.setItem(KUNCI_TOKEN, t)
    else localStorage.removeItem(KUNCI_TOKEN)
  } catch {
    /* abaikan */
  }
}

export class GalatApi extends Error {
  status: number
  constructor(status: number, pesan: string) {
    super(pesan)
    this.status = status
  }
}

export async function api<T>(jalur: string, opsi: RequestInit = {}): Promise<T> {
  const headers = new Headers(opsi.headers)
  headers.set('Accept', 'application/json')
  if (opsi.body && !headers.has('Content-Type')) headers.set('Content-Type', 'application/json')
  const token = ambilToken()
  if (token) headers.set('Authorization', `Bearer ${token}`)

  // Cache-buster untuk GET supaya data selalu segar (hindari cache peramban).
  const metode = (opsi.method ?? 'GET').toUpperCase()
  const url = metode === 'GET' ? `${BASE}/${jalur}${jalur.includes('?') ? '&' : '?'}_=${Date.now()}` : `${BASE}/${jalur}`
  const res = await fetch(url, { cache: 'no-store', ...opsi, headers })
  if (res.status === 204) return undefined as T

  let data: any = null
  try {
    data = await res.json()
  } catch {
    /* kosong */
  }

  if (!res.ok) {
    throw new GalatApi(res.status, data?.message ?? 'Terjadi kesalahan. Coba lagi.')
  }
  return data as T
}
