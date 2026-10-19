import { createRouter, createWebHashHistory } from 'vue-router'
import { useAuth } from '@/stores/auth'

const routes = [
  { path: '/login', name: 'login', component: () => import('@/pages/Login.vue'), meta: { publik: true, polos: true } },
  {
    path: '/',
    component: () => import('@/layouts/AppLayout.vue'),
    children: [
      { path: 'pencarian', component: () => import('@/pages/Pencarian.vue'), meta: { judul: 'Pencarian' } },
      { path: 'dokumen/baru', component: () => import('@/pages/DokumenDetail.vue'), meta: { judul: 'Tambah Dokumen', admin: true } },
      { path: 'dokumen/:id', component: () => import('@/pages/DokumenDetail.vue'), meta: { judul: 'Butir Acuan Dokumen' } },
      { path: 'pengaturan/kelola/:bagian', component: () => import('@/pages/Administrasi.vue'), meta: { judul: 'Administrasi', admin: true } },
      { path: '', name: 'dashboard', component: () => import('@/pages/Dashboard.vue'), meta: { judul: 'Dasbor' } },
      { path: 'library', name: 'library', component: () => import('@/pages/CpLibrary.vue'), meta: { judul: 'Pustaka CP' } },
      { path: 'library/:kode', name: 'cp-detail', component: () => import('@/pages/CpDetail.vue'), meta: { judul: 'Detail CP' } },
      { path: 'approval', name: 'approval', component: () => import('@/pages/Approval.vue'), meta: { judul: 'Pengesahan' } },
      { path: 'cp-aktif', name: 'cp-aktif', component: () => import('@/pages/CpAktif.vue'), meta: { judul: 'CP Aktif' } },
      { path: 'evaluasi', name: 'evaluasi', component: () => import('@/pages/Evaluasi.vue'), meta: { judul: 'Evaluasi CP' } },
      { path: 'dokumen', name: 'dokumen', component: () => import('@/pages/Dokumen.vue'), meta: { judul: 'Dokumen Panduan' } },
      { path: 'padanan/kptl', name: 'kptl', component: () => import('@/pages/Padanan.vue'), meta: { judul: 'Padanan KPTL' } },
      { path: 'padanan/snomed', name: 'snomed', component: () => import('@/pages/Padanan.vue'), meta: { judul: 'Padanan SNOMED-CT' } },
      { path: 'pengaturan', name: 'pengaturan', component: () => import('@/pages/Pengaturan.vue'), meta: { judul: 'Pengaturan', admin: true } },
      {
        path: 'pengaturan/panduan',
        name: 'panduan',
        component: () => import('@/pages/PanduanSuperAdmin.vue'),
        meta: { judul: 'Panduan Super Admin', superAdmin: true },
      },
    ],
  },
  { path: '/:pathMatch(.*)*', redirect: '/' },
]

export const router = createRouter({
  history: createWebHashHistory(),
  routes,
  scrollBehavior: () => ({ top: 0 }),
})

router.beforeEach(async (to) => {
  const auth = useAuth()
  if (!auth.siap) await auth.muatProfil()

  if (!to.meta.publik && !auth.masuk) {
    return { name: 'login', query: to.fullPath !== '/' ? { ke: to.fullPath } : undefined }
  }
  if (to.name === 'login' && auth.masuk) return { path: '/' }
  if (to.meta.admin && !auth.isAdmin) return { path: '/' }
  // Kosmetik saja — batas sesungguhnya: backend membalas 403 untuk selain SUPER_ADMIN.
  if (to.meta.superAdmin && !auth.isSuperAdmin) return { path: auth.isAdmin ? '/pengaturan' : '/' }
})
