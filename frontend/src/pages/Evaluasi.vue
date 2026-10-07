<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { TrendingUp, ChevronDown, TriangleAlert, CircleCheck } from 'lucide-vue-next'
import { api, GalatApi } from '@/lib/api'
import { angka, rupiahRingkas } from '@/lib/format'
import PageHeader from '@/components/PageHeader.vue'

interface Row {
  kode: string
  nama: string
  kelompok_diagnosis: string
  aktif_sejak: string
  episode_sesudah: number
  perlu_ditinjau: boolean
  temuan: string[]
}
interface Mutu {
  severity: string
  episode: number
  los_median: number
  los_p75: number
  target_los: number
  target_p75: number
  persen_lewat_p75: number
  episode_icu: number
}
interface Detail extends Row {
  kendali_mutu: Mutu[]
  kendali_biaya: {
    baur_severity_sebelum: Record<string, number>
    baur_severity_sesudah: Record<string, number>
    tarif_rata_rata: number
    total_klaim: number
    persen_diagnosis_di_luar_pola: number
  }
}

const ringkasan = ref<Record<string, any> | null>(null)
const rows = ref<Row[]>([])
const galat = ref('')
const expanded = ref('')
const detail = ref<Record<string, Detail>>({})
const memuatDetail = ref('')

onMounted(async () => {
  try {
    const r = await api<{ ringkasan: Record<string, any>; data: Row[] }>('evaluasi')
    ringkasan.value = r.ringkasan
    rows.value = r.data
  } catch (e) {
    galat.value = e instanceof GalatApi ? e.message : 'Gagal memuat evaluasi.'
  }
})

async function toggle(kode: string) {
  if (expanded.value === kode) {
    expanded.value = ''
    return
  }
  expanded.value = kode
  if (!detail.value[kode]) {
    memuatDetail.value = kode
    try {
      detail.value[kode] = await api<Detail>(`evaluasi/${kode}`)
    } catch {
      /* abaikan */
    } finally {
      memuatDetail.value = ''
    }
  }
}
</script>

<template>
  <div>
    <PageHeader judul="Evaluasi CP" demo sub="Kepatuhan CP aktif terhadap pelaksanaan nyata — membandingkan klaim sesudah CP disahkan dengan target saat pengesahan." />

    <div v-if="galat" class="galat-box">{{ galat }}</div>

    <template v-if="ringkasan">
      <div class="ringkas">
        <div class="rk card"><p class="rk-n angka">{{ ringkasan.cp_aktif }}</p><p class="rk-l">CP aktif</p></div>
        <div class="rk card"><p class="rk-n angka" style="color: var(--hijau)">{{ ringkasan.bisa_dievaluasi }}</p><p class="rk-l">Bisa dievaluasi</p></div>
        <div class="rk card"><p class="rk-n angka" style="color: var(--merah)">{{ ringkasan.perlu_ditinjau }}</p><p class="rk-l">Perlu ditinjau</p></div>
        <div class="rk card"><p class="rk-n angka" style="color: var(--text-3)">{{ ringkasan.belum_aktif }}</p><p class="rk-l">Belum aktif</p></div>
        <div class="rk card"><p class="rk-n kecil">{{ ringkasan.klaim_sampai }}</p><p class="rk-l">Klaim s.d.</p></div>
      </div>

      <div class="card panel">
        <ul class="list">
          <li v-for="r in rows" :key="r.kode" class="item-wrap">
            <button class="item" :class="{ buka: expanded === r.kode }" @click="toggle(r.kode)">
              <ChevronDown :size="16" class="chev" :class="{ putar: expanded === r.kode }" />
              <span class="mono kode">{{ r.kode }}</span>
              <div class="isi">
                <p class="nama">{{ r.nama }}</p>
                <p class="meta">{{ r.kelompok_diagnosis }} · aktif sejak {{ r.aktif_sejak }} · {{ angka(r.episode_sesudah) }} episode sesudah</p>
              </div>
              <span v-if="r.perlu_ditinjau" class="badge hambat"><TriangleAlert :size="12" /> {{ r.temuan.length }} temuan</span>
              <span v-else class="badge ok"><CircleCheck :size="12" /> Sesuai target</span>
            </button>

            <div v-if="expanded === r.kode" class="detail-panel">
              <div v-if="memuatDetail === r.kode" class="memuat-kecil">Memuat…</div>
              <template v-else-if="detail[r.kode]">
                <!-- Temuan -->
                <div v-if="detail[r.kode].temuan.length" class="temuan">
                  <p class="d-judul">Temuan</p>
                  <ul>
                    <li v-for="(t, i) in detail[r.kode].temuan" :key="i"><TriangleAlert :size="13" /> {{ t }}</li>
                  </ul>
                </div>

                <!-- Kendali mutu -->
                <p class="d-judul">Kendali mutu — LOS per severity</p>
                <table class="d-tabel">
                  <thead><tr><th>Sev</th><th class="kanan">Episode</th><th class="kanan">LOS median</th><th class="kanan">Target (p75)</th><th class="kanan">% lewat p75</th><th class="kanan">ICU</th></tr></thead>
                  <tbody>
                    <tr v-for="m in detail[r.kode].kendali_mutu" :key="m.severity">
                      <td><span class="mono">{{ m.severity }}</span></td>
                      <td class="kanan angka">{{ angka(m.episode) }}</td>
                      <td class="kanan angka" :class="{ lewat: m.los_median > m.target_p75 }">{{ m.los_median }} hari</td>
                      <td class="kanan angka muted">{{ m.target_los }} / {{ m.target_p75 }}</td>
                      <td class="kanan angka" :class="{ lewat: m.persen_lewat_p75 >= 30 }">{{ m.persen_lewat_p75 }}%</td>
                      <td class="kanan angka muted">{{ m.episode_icu }}</td>
                    </tr>
                  </tbody>
                </table>

                <!-- Kendali biaya -->
                <p class="d-judul mt">Kendali biaya (tarif INA-CBG)</p>
                <div class="biaya">
                  <div class="b-kotak">
                    <p class="b-l">Tarif rata-rata</p>
                    <p class="b-n angka">{{ rupiahRingkas(detail[r.kode].kendali_biaya.tarif_rata_rata) }}</p>
                  </div>
                  <div class="b-kotak">
                    <p class="b-l">Total klaim sesudah</p>
                    <p class="b-n angka">{{ rupiahRingkas(detail[r.kode].kendali_biaya.total_klaim) }}</p>
                  </div>
                  <div class="b-kotak">
                    <p class="b-l">Diagnosis di luar pola</p>
                    <p class="b-n angka">{{ detail[r.kode].kendali_biaya.persen_diagnosis_di_luar_pola }}%</p>
                  </div>
                  <div class="b-kotak lebar">
                    <p class="b-l">Bauran severity (sebelum → sesudah)</p>
                    <div class="baur">
                      <span v-for="sv in ['I', 'II', 'III']" :key="sv" class="baur-item">
                        <b>{{ sv }}</b> {{ detail[r.kode].kendali_biaya.baur_severity_sebelum[sv] }}% → {{ detail[r.kode].kendali_biaya.baur_severity_sesudah[sv] }}%
                      </span>
                    </div>
                  </div>
                </div>
              </template>
            </div>
          </li>
        </ul>
      </div>
    </template>
  </div>
</template>

<style scoped>
.galat-box {
  padding: 14px 16px;
  border-radius: var(--radius);
  background: var(--merah-bg);
  color: var(--merah);
  font-weight: 600;
}
.ringkas {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(150px, 1fr));
  gap: 14px;
  margin-bottom: 16px;
}
.rk {
  padding: 16px 18px;
}
.rk-n {
  font-family: var(--font-judul);
  font-size: 26px;
  font-weight: 800;
}
.rk-n.kecil {
  font-size: 16px;
  margin-top: 8px;
}
.rk-l {
  font-size: 12.5px;
  color: var(--text-2);
  margin-top: 2px;
}
.panel {
  padding: 8px;
}
.list {
  list-style: none;
  padding: 0;
  margin: 0;
}
.item {
  width: 100%;
  display: flex;
  align-items: center;
  gap: 14px;
  padding: 14px;
  background: transparent;
  border: none;
  border-radius: var(--radius-sm);
  text-align: left;
}
.item:hover {
  background: var(--bg);
}
.item.buka {
  background: var(--bg);
}
.chev {
  color: var(--text-3);
  transition: transform 0.18s;
}
.chev.putar {
  transform: rotate(180deg);
}
.kode {
  font-size: 12.5px;
  font-weight: 700;
  color: var(--brand-aktif);
  width: 62px;
  flex-shrink: 0;
}
.isi {
  flex: 1;
  min-width: 0;
}
.nama {
  font-weight: 600;
  font-size: 14px;
}
.meta {
  font-size: 12px;
  color: var(--text-3);
  margin-top: 2px;
}
.hambat {
  background: var(--merah-bg);
  color: var(--merah);
}
.ok {
  background: var(--hijau-bg);
  color: var(--hijau);
}
.detail-panel {
  padding: 8px 16px 20px 44px;
}
.memuat-kecil {
  padding: 20px;
  color: var(--text-3);
}
.d-judul {
  font-family: var(--font-judul);
  font-weight: 700;
  font-size: 13px;
  margin-bottom: 8px;
}
.d-judul.mt {
  margin-top: 18px;
}
.temuan {
  background: var(--merah-bg);
  border-radius: var(--radius-sm);
  padding: 12px 14px;
  margin-bottom: 16px;
}
.temuan ul {
  list-style: none;
  padding: 0;
  margin: 0;
  display: flex;
  flex-direction: column;
  gap: 6px;
}
.temuan li {
  display: flex;
  align-items: flex-start;
  gap: 8px;
  font-size: 13px;
  color: var(--merah);
}
.d-tabel {
  width: 100%;
  border-collapse: collapse;
  font-size: 13px;
}
.d-tabel th {
  text-align: left;
  font-size: 10.5px;
  font-weight: 700;
  letter-spacing: 0.5px;
  text-transform: uppercase;
  color: var(--text-3);
  padding: 7px 10px;
  border-bottom: 1px solid var(--border);
}
.d-tabel td {
  padding: 9px 10px;
  border-bottom: 1px solid var(--border);
}
.d-tabel tbody tr:last-child td {
  border-bottom: none;
}
.kanan {
  text-align: right;
}
.lewat {
  color: var(--merah);
  font-weight: 700;
}
.biaya {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 12px;
}
.b-kotak {
  background: var(--bg);
  border: 1px solid var(--border);
  border-radius: var(--radius-sm);
  padding: 12px 14px;
}
.b-kotak.lebar {
  grid-column: 1 / -1;
}
.b-l {
  font-size: 11.5px;
  color: var(--text-3);
  text-transform: uppercase;
  letter-spacing: 0.4px;
}
.b-n {
  font-family: var(--font-judul);
  font-weight: 800;
  font-size: 18px;
  margin-top: 4px;
}
.baur {
  display: flex;
  gap: 18px;
  margin-top: 8px;
  flex-wrap: wrap;
}
.baur-item {
  font-size: 13px;
  color: var(--text-2);
}
.baur-item b {
  color: var(--brand-aktif);
  margin-right: 4px;
}
@media (max-width: 720px) {
  .biaya {
    grid-template-columns: 1fr;
  }
}
</style>
