<script setup lang="ts">
import { reactive, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { Mail, Lock, Eye, EyeOff, ArrowRight, Network, ShieldCheck, Sparkles, CircleAlert } from 'lucide-vue-next'
import { useAuth } from '@/stores/auth'
import { GalatApi } from '@/lib/api'
import LogoSpkd from '@/components/LogoSpkd.vue'

const auth = useAuth()
const route = useRoute()
const router = useRouter()

const form = reactive({ email: '', password: '' })
const lihat = ref(false)
const memuat = ref(false)
const galat = ref('')

async function kirim() {
  if (!form.email || !form.password) return
  galat.value = ''
  memuat.value = true
  try {
    await auth.login(form.email.trim().toLowerCase(), form.password)
    const ke = typeof route.query.ke === 'string' ? route.query.ke : '/'
    router.push(ke)
  } catch (e) {
    galat.value = e instanceof GalatApi ? e.message : 'Gagal terhubung ke server.'
  } finally {
    memuat.value = false
  }
}

const akunDemo = [
  { label: 'Super Admin', email: 'super_admin@demo.local' },
  { label: 'Tim CP', email: 'tim_cp@demo.local' },
  { label: 'DPJP', email: 'dpjp@demo.local' },
  { label: 'Komite Medik', email: 'komite_medik@demo.local' },
  { label: 'Direktur', email: 'direktur@demo.local' },
]
function isiDemo(email: string) {
  form.email = email
  form.password = 'Demo-SPKD-2026'
  kirim()
}

const pilar = [
  { icon: Network, teks: 'CP disusun dari pola klaim INA-CBG riil & acuan PNPK/PPK' },
  { icon: ShieldCheck, teks: 'Pengesahan 4 tahap dengan audit trail penuh, tanpa data pasien' },
  { icon: Sparkles, teks: 'Dari data klaim ke CP aktif yang siap memandu pelayanan' },
]
const statistik = [
  { n: '4', l: 'Tahap pengesahan' },
  { n: '15', l: 'Peran RBAC' },
  { n: 'INA-CBG', l: 'Basis data klaim' },
  { n: '100%', l: 'Audit trail' },
]
</script>

<template>
  <div class="login">
    <!-- Panel merek -->
    <section class="merek">
      <div class="merek-head"><LogoSpkd terang /></div>

      <div class="merek-isi">
        <h1>From Hospital Data<br />to Clinical Pathway.</h1>
        <p class="deskripsi">
          Clinical Pathway Generator PT SPKD — menyusun template Clinical Pathway dari data klaim INA-CBG,
          menetapkan acuan & rencana klinis, lalu mengesahkannya dari Tim CP hingga Direktur.
        </p>
        <ul class="pilar">
          <li v-for="(p, i) in pilar" :key="i">
            <span class="pilar-ikon"><component :is="p.icon" :size="18" /></span>
            {{ p.teks }}
          </li>
        </ul>
      </div>

      <div class="statistik">
        <div v-for="s in statistik" :key="s.l">
          <p class="s-n">{{ s.n }}</p>
          <p class="s-l">{{ s.l }}</p>
        </div>
      </div>

      <div class="glow glow-a" />
      <div class="glow glow-b" />
    </section>

    <!-- Formulir -->
    <section class="form-panel">
      <div class="form-box">
        <LogoSpkd class="mobile-logo" />
        <h2>Masuk ke MedPath</h2>
        <p class="sub">Gunakan akun yang diberikan Admin RS.</p>

        <form @submit.prevent="kirim">
          <label class="field">
            <span class="field-label">Email</span>
            <span class="input">
              <Mail :size="17" class="input-ikon" />
              <input v-model="form.email" type="email" autocomplete="username" placeholder="nama@rumahsakit.go.id" />
            </span>
          </label>

          <label class="field">
            <span class="field-label">Kata sandi</span>
            <span class="input">
              <Lock :size="17" class="input-ikon" />
              <input v-model="form.password" :type="lihat ? 'text' : 'password'" autocomplete="current-password" placeholder="••••••••••" />
              <button type="button" class="mata" @click="lihat = !lihat">
                <EyeOff v-if="lihat" :size="17" />
                <Eye v-else :size="17" />
              </button>
            </span>
          </label>

          <div v-if="galat" class="galat">
            <CircleAlert :size="16" />
            {{ galat }}
          </div>

          <button type="submit" class="btn btn-brand masuk" :disabled="memuat || !form.email || !form.password">
            <span v-if="memuat">Memproses…</span>
            <template v-else>Masuk <ArrowRight :size="17" /></template>
          </button>
        </form>

        <p class="demo-ket">Akun demo (klik untuk masuk) — sandi <strong>Demo-SPKD-2026</strong></p>
        <div class="chips">
          <button v-for="a in akunDemo" :key="a.email" class="chip" @click="isiDemo(a.email)">{{ a.label }}</button>
        </div>
      </div>
    </section>
  </div>
</template>

<style scoped>
.login {
  min-height: 100dvh;
  display: grid;
  grid-template-columns: 1.05fr 1fr;
}
/* Panel merek */
.merek {
  position: relative;
  overflow: hidden;
  padding: 44px 56px;
  display: flex;
  flex-direction: column;
  justify-content: space-between;
  color: #fff;
  background: var(--gradient-hero);
}
.merek-isi {
  position: relative;
  z-index: 2;
  max-width: 540px;
}
.merek-isi h1 {
  font-size: 42px;
  font-weight: 800;
  line-height: 1.12;
  letter-spacing: -1px;
  color: #fff;
}
.deskripsi {
  margin-top: 20px;
  color: rgba(255, 255, 255, 0.72);
  font-size: 15px;
  line-height: 1.65;
}
.pilar {
  list-style: none;
  padding: 0;
  margin-top: 30px;
  display: flex;
  flex-direction: column;
  gap: 16px;
}
.pilar li {
  display: flex;
  align-items: center;
  gap: 14px;
  font-size: 14px;
  color: rgba(255, 255, 255, 0.86);
}
.pilar-ikon {
  width: 36px;
  height: 36px;
  flex-shrink: 0;
  border-radius: 10px;
  display: grid;
  place-items: center;
  color: var(--brand);
  background: rgba(242, 105, 94, 0.14);
  border: 1px solid rgba(242, 105, 94, 0.22);
}
.statistik {
  position: relative;
  z-index: 2;
  display: flex;
  gap: 42px;
  flex-wrap: wrap;
}
.s-n {
  font-family: var(--font-judul);
  font-size: 26px;
  font-weight: 800;
  color: var(--brand);
}
.s-l {
  font-size: 12px;
  color: rgba(255, 255, 255, 0.6);
  margin-top: 2px;
}
.glow {
  position: absolute;
  border-radius: 50%;
  pointer-events: none;
}
.glow-a {
  width: 460px;
  height: 460px;
  right: -140px;
  top: -120px;
  border: 1px solid rgba(242, 105, 94, 0.12);
}
.glow-b {
  width: 640px;
  height: 640px;
  left: -40px;
  bottom: -320px;
  border: 1px solid rgba(242, 105, 94, 0.08);
}
/* Panel form */
.form-panel {
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 40px;
  background: var(--bg);
}
.form-box {
  width: 100%;
  max-width: 400px;
}
.mobile-logo {
  display: none;
  margin-bottom: 28px;
}
.form-box h2 {
  font-size: 26px;
  font-weight: 800;
}
.sub {
  margin-top: 6px;
  color: var(--text-2);
  font-size: 14px;
}
form {
  margin-top: 26px;
  display: flex;
  flex-direction: column;
  gap: 16px;
}
.field-label {
  display: block;
  font-size: 13px;
  font-weight: 700;
  margin-bottom: 7px;
  color: var(--text);
}
.input {
  display: flex;
  align-items: center;
  gap: 10px;
  height: 48px;
  padding: 0 12px;
  background: var(--surface);
  border: 1px solid var(--border-kuat);
  border-radius: var(--radius-sm);
  transition: border-color 0.15s, box-shadow 0.15s;
}
.input:focus-within {
  border-color: var(--brand);
  box-shadow: 0 0 0 3px rgba(242, 105, 94, 0.18);
}
.input-ikon {
  color: var(--text-3);
  flex-shrink: 0;
}
.input input {
  flex: 1;
  border: none;
  outline: none;
  background: transparent;
  font-size: 14.5px;
}
.mata {
  border: none;
  background: transparent;
  color: var(--text-3);
  display: grid;
  place-items: center;
}
.masuk {
  height: 48px;
  margin-top: 4px;
  font-size: 15px;
}
.masuk:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}
.galat {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 10px 12px;
  border-radius: var(--radius-sm);
  background: var(--merah-bg);
  color: var(--merah);
  font-size: 13px;
  font-weight: 600;
}
.demo-ket {
  margin-top: 26px;
  font-size: 12px;
  color: var(--text-3);
  text-align: center;
}
.chips {
  margin-top: 12px;
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  justify-content: center;
}
.chip {
  padding: 7px 14px;
  border-radius: var(--radius-pill);
  border: 1px solid var(--border-kuat);
  background: var(--surface);
  font-size: 12.5px;
  font-weight: 600;
  color: var(--text-2);
  transition: border-color 0.14s, color 0.14s, background 0.14s;
}
.chip:hover {
  border-color: var(--brand);
  color: var(--brand-gelap);
  background: var(--brand-light);
}
@media (max-width: 920px) {
  .login {
    grid-template-columns: 1fr;
  }
  .merek {
    display: none;
  }
  .mobile-logo {
    display: flex;
  }
}
</style>
