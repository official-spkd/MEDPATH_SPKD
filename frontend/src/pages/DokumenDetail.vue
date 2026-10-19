<script setup lang="ts">
import {ref,watch} from 'vue'
import {useRoute,useRouter} from 'vue-router'
import {api,GalatApi} from '@/lib/api'
import {LABEL_SUMBER} from '@/lib/format'
import PageHeader from '@/components/PageHeader.vue'
interface Dokumen {id:number;nama:string;sumber:string;penerbit:string;tahun:number;icd10_cakupan:string[];butir_count:number;butir:string[]|null}
const route=useRoute(),router=useRouter(),d=ref<Dokumen|null>(null),galat=ref(''),proses=ref(false)
const form=ref({nama:'',sumber:'PNPK',penerbit:'',tahun:new Date().getFullYear(),cakupan:'',butir:''})
watch(()=>route.params.id,async(id)=>{d.value=null;galat.value='';if(!id)return;try{d.value=await api<Dokumen>(`dokumen-panduan/${id}`)}catch(e){galat.value=e instanceof GalatApi?e.message:'Gagal memuat dokumen.'}},{immediate:true})
async function simpan(){proses.value=true;galat.value='';try{const r=await api<Dokumen>('dokumen-panduan',{method:'POST',body:JSON.stringify({...form.value,icd10_cakupan:form.value.cakupan.split(',').map(x=>x.trim()).filter(Boolean),butir:form.value.butir.split('\n').filter(x=>x.trim())})});router.replace(`/dokumen/${r.id}`)}catch(e){galat.value=e instanceof GalatApi?e.message:'Gagal menyimpan dokumen.'}finally{proses.value=false}}
</script>
<template><div><PageHeader :judul="route.params.id ? (d?.nama ?? 'Butir acuan dokumen') : 'Tambah Dokumen'"/><router-link to="/dokumen">← Daftar dokumen</router-link><p v-if="galat" role="alert">{{galat}}</p>
<form v-if="!route.params.id" class="card panel" @submit.prevent="simpan"><label>Nama dokumen<input v-model="form.nama" required/></label><label>Sumber<select v-model="form.sumber"><option v-for="k in ['PNPK','PPK_RS','PPK_ASOSIASI']" :key="k" :value="k">{{LABEL_SUMBER[k]}}</option></select></label><label>Penerbit<input v-model="form.penerbit" required/></label><label>Tahun<input v-model.number="form.tahun" type="number" min="1900" :max="new Date().getFullYear()+1" required/></label><label>Cakupan ICD-10 (pisahkan dengan koma)<input v-model="form.cakupan"/></label><label>Butir acuan dari dokumen sumber (satu butir per baris)<textarea v-model="form.butir" rows="6"/></label><button class="btn btn-brand" :disabled="proses">Simpan dokumen</button></form>
<section v-else-if="d" class="card panel"><p>{{LABEL_SUMBER[d.sumber] ?? d.sumber}} · {{d.penerbit}} · {{d.tahun}}</p><h2>Butir acuan ({{d.butir?.length ?? 0}})</h2><ol v-if="d.butir?.length"><li v-for="(b,i) in d.butir" :key="i">{{b}}</li></ol><p v-else>Isi butir dokumen ini belum tersedia. Data awal hanya memuat metadata dan jumlah {{d.butir_count}} butir; teks sumber belum diunggah.</p><p>Cakupan: {{d.icd10_cakupan?.join(', ') || 'Belum ditetapkan'}}</p></section>
</div></template>
<style scoped>.panel{padding:24px;margin-top:20px}form{max-width:760px}label{display:grid;gap:6px;margin-bottom:16px}input,select,textarea{padding:10px;border:1px solid var(--border);border-radius:8px;background:var(--surface);color:var(--text);width:100%}li{margin:12px 0}ol{padding-left:24px}h2{margin-top:20px}</style>
