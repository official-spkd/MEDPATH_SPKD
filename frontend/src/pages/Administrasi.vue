<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { api, GalatApi } from '@/lib/api'
import { useAuth } from '@/stores/auth'
import PageHeader from '@/components/PageHeader.vue'
interface Pengguna { id:number; nama:string; email:string; peran:string; aktif:boolean }
interface Profil { nama:string; tipe:string; kepemilikan:string; regional:string; kelas:string }
interface Kunci { id:string; nama:string; dibuat:string; aktif:boolean }
interface Data { pengguna:Pengguna[]; profil:Profil; kunci:Kunci[]; demo_dinonaktifkan:boolean }
const route=useRoute(), auth=useAuth()
const bagian=computed(()=>String(route.params.bagian))
const judul=computed(()=>({pengguna:'Pengguna & Peran',profil:'Profil Rumah Sakit',integrasi:'Kunci Integrasi'}[bagian.value] ?? 'Pengaturan'))
const data=ref<Data|null>(null), pesan=ref(''), proses=ref(false), namaKunci=ref(''), kunciBaru=ref('')
const peran=['SUPER_ADMIN','SYSTEM_ADMIN','ADMIN_RS','KOORDINATOR_CP','TIM_CP','DPJP','DOKTER','APOTEKER','KOMITE_MEDIK','DIREKTUR']
const penggunaBaru=ref({nama:'',email:'',peran:'TIM_CP',password:''})
async function tambahPengguna(){proses.value=true;pesan.value='';try{data.value=await api<Data>('admin/pengguna',{method:'POST',body:JSON.stringify(penggunaBaru.value)});penggunaBaru.value={nama:'',email:'',peran:'TIM_CP',password:''};pesan.value='Akun tersimpan. Masuk dengan akun tersebut untuk memverifikasi akses.'}catch(e){pesan.value=e instanceof GalatApi?e.message:'Gagal membuat akun.'}finally{proses.value=false}}
async function nonaktifkanDemo(){if(!confirm('Nonaktifkan semua akun demo? Akun Super Admin non-demo yang sedang dipakai tetap aktif.'))return;try{data.value=await api<Data>('admin/nonaktifkan-demo',{method:'POST'});pesan.value='Akun demo dinonaktifkan.'}catch(e){pesan.value=e instanceof GalatApi?e.message:'Gagal menonaktifkan demo.'}}
onMounted(async()=>{try {data.value=await api<Data>('admin')}catch(e){pesan.value=e instanceof GalatApi?e.message:'Gagal memuat pengaturan.'}})
async function simpan(path:string,body:unknown) {
 proses.value=true;pesan.value=''
 try{data.value=await api<Data>(path,{method:'PUT',body:JSON.stringify(body)});pesan.value='Perubahan tersimpan.'}catch(e){pesan.value=e instanceof GalatApi?e.message:'Gagal menyimpan.'}finally{proses.value=false}
}
async function terbitkan() {
 proses.value=true;pesan.value='';kunciBaru.value=''
 try{const r=await api<{kunci:string}>('admin/kunci',{method:'POST',body:JSON.stringify({nama:namaKunci.value})});kunciBaru.value=r.kunci;data.value=await api<Data>('admin');namaKunci.value=''}catch(e){pesan.value=e instanceof GalatApi?e.message:'Gagal menerbitkan kunci.'}finally{proses.value=false}
}
async function cabut(k:Kunci){if(!confirm(`Cabut akses integrasi ${k.nama}?`))return;try{data.value=await api<Data>(`admin/kunci/${k.id}`,{method:'DELETE'})}catch(e){pesan.value=e instanceof GalatApi?e.message:'Gagal mencabut kunci.'}}
</script>
<template><div><PageHeader :judul="judul" sub="Pengelolaan administrasi MedPath."/><router-link to="/pengaturan">← Kembali ke pengaturan</router-link>
<p v-if="pesan" role="status" class="pesan">{{ pesan }}</p>
<div v-if="data" class="card panel">
<template v-if="bagian==='pengguna'"><form v-if="auth.isSuperAdmin" @submit.prevent="tambahPengguna"><h2>Tambah pengguna non-demo</h2><label>Nama<input v-model="penggunaBaru.nama" required/></label><label>Email<input v-model="penggunaBaru.email" type="email" autocomplete="off" required/></label><label>Peran<select v-model="penggunaBaru.peran"><option v-for="p in peran" :key="p">{{p}}</option></select></label><label>Kata sandi<input v-model="penggunaBaru.password" type="password" minlength="12" autocomplete="new-password" required/></label><button class="btn btn-brand" :disabled="proses">Buat pengguna</button></form><div v-if="auth.isSuperAdmin && !data.demo_dinonaktifkan" class="pesan"><p>Akun demo masih aktif. Buat akun Super Admin non-demo, masuk dengan akun tersebut, lalu nonaktifkan akun demo.</p><button class="btn btn-outline" :disabled="auth.pengguna?.email.endsWith('@demo.local')" @click="nonaktifkanDemo">Nonaktifkan akun demo</button></div><div class="tabel-bungkus"><table class="tabel"><thead><tr><th>Nama</th><th>Email</th><th>Peran</th><th>Aktif</th><th></th></tr></thead><tbody><tr v-for="u in data.pengguna" :key="u.id"><td><input v-model="u.nama" aria-label="Nama pengguna" :disabled="!auth.isSuperAdmin || u.id===auth.pengguna?.id"/></td><td>{{u.email}}</td><td><select v-model="u.peran" aria-label="Peran pengguna" :disabled="!auth.isSuperAdmin || u.id===auth.pengguna?.id"><option v-for="p in peran" :key="p">{{p}}</option></select></td><td><input v-model="u.aktif" type="checkbox" aria-label="Akun aktif" :disabled="!auth.isSuperAdmin || u.id===auth.pengguna?.id"/></td><td><button v-if="auth.isSuperAdmin && u.id!==auth.pengguna?.id" class="btn btn-brand" :disabled="proses" @click="simpan('admin/pengguna',u)">Simpan</button></td></tr></tbody></table></div></template>
<form v-else-if="bagian==='profil'" @submit.prevent="simpan('admin/profil',data.profil)"><label>Nama rumah sakit<input v-model="data.profil.nama" required/></label><label>Tipe RS<select v-model="data.profil.tipe"><option v-for="v in ['A','B','C','D']" :key="v">{{v}}</option></select></label><label>Kepemilikan<input v-model="data.profil.kepemilikan" required/></label><label>Regional<input v-model="data.profil.regional" required/></label><label>Kelas tertinggi<input v-model="data.profil.kelas" required/></label><button class="btn btn-brand" :disabled="proses">Simpan profil</button></form>
<template v-else-if="bagian==='integrasi'"><p>Kunci hanya mengizinkan pembacaan CP aktif melalui <code>/api/v1/integrasi/cp-aktif</code>.</p><form v-if="auth.isSuperAdmin" @submit.prevent="terbitkan"><label>Nama integrasi<input v-model="namaKunci" required placeholder="Contoh: TransCPR-X QA"/></label><button class="btn btn-brand" :disabled="proses">Terbitkan kunci</button></form><div v-if="kunciBaru" class="pesan"><p>Salin kunci ini sekarang. Kunci tidak ditampilkan lagi setelah halaman ditutup.</p><input :value="kunciBaru" readonly aria-label="Kunci integrasi baru"/></div><ul><li v-for="k in data.kunci" :key="k.id">{{k.nama}} · {{k.aktif?'Aktif':'Dicabut'}} · {{k.dibuat}} <button v-if="k.aktif && auth.isSuperAdmin" class="btn btn-outline" @click="cabut(k)">Cabut</button></li></ul><p v-if="!data.kunci.length">Belum ada kunci integrasi.</p></template>
</div></div></template>
<style scoped>.panel{padding:20px;margin-top:18px;overflow:auto}form{max-width:600px}label{display:grid;gap:6px;margin:12px 0}input,select{padding:10px;border:1px solid var(--border);border-radius:8px;background:var(--surface);color:var(--text);max-width:100%}.pesan{padding:14px;background:var(--brand-light);margin:12px 0;overflow-wrap:anywhere}li{padding:12px 0}.tabel-bungkus{overflow:auto}</style>
